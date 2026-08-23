package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-jc-challange2/config"
	"go-jc-challange2/models"
	"go-jc-challange2/utils"
)

// ============================================================================
// CONTROLLER PENGAJUAN CUTI
// ============================================================================

// CreateLeave menangani POST /api/leaves, yaitu pengajuan cuti oleh karyawan.
// Setiap pengajuan baru selalu berstatus PENDING dan menunggu keputusan HR.
func CreateLeave(c *gin.Context) {
	var req models.LeaveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// --- Validasi relasi: karyawan harus ada --------------------------------
	var employee models.Employee
	if err := config.DB.First(&employee, req.EmployeeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusBadRequest, "Karyawan dengan ID tersebut tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data karyawan")
		return
	}

	if employee.Status != models.EmployeeStatusActive {
		utils.Error(c, http.StatusUnprocessableEntity,
			"Karyawan berstatus "+employee.Status+" tidak dapat mengajukan cuti")
		return
	}

	// --- Hitung durasi cuti (inklusif) --------------------------------------
	// CountDaysInclusive sekaligus memvalidasi bahwa tanggal selesai tidak
	// mendahului tanggal mulai, sebuah aturan antar-field yang tidak bisa
	// diperiksa oleh tag binding biasa.
	totalDays, err := utils.CountDaysInclusive(req.StartDate, req.EndDate)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, utils.SentenceCase(err.Error()))
		return
	}

	// --- Cegah pengajuan yang bertabrakan dengan cuti lain -------------------
	// Dua rentang dianggap beririsan bila start_date <= EndDate baru
	// DAN end_date >= StartDate baru. Cuti yang sudah ditolak diabaikan.
	var overlapping int64
	err = config.DB.Model(&models.Leave{}).
		Where("employee_id = ? AND status <> ?", req.EmployeeID, models.LeaveStatusRejected).
		Where("start_date <= ? AND end_date >= ?", req.EndDate, req.StartDate).
		Count(&overlapping).Error
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa jadwal cuti")
		return
	}
	if overlapping > 0 {
		utils.Error(c, http.StatusConflict,
			"Karyawan sudah memiliki pengajuan cuti pada rentang tanggal tersebut")
		return
	}

	leave := models.Leave{
		EmployeeID: req.EmployeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TotalDays:  totalDays,
		Reason:     strings.TrimSpace(req.Reason),
		Status:     models.LeaveStatusPending,
	}

	if err := config.DB.Create(&leave).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan pengajuan cuti", err.Error())
		return
	}

	config.DB.Preload("Employee").First(&leave, leave.ID)

	utils.SuccessWithMeta(c, http.StatusCreated,
		"Pengajuan cuti berhasil dibuat dan menunggu persetujuan HR", leave, gin.H{
			"sisa_cuti_saat_ini": employee.LeaveBalance,
			"lama_cuti_diajukan": totalDays,
		})
}

// GetLeaves menangani GET /api/leaves.
//
// Filter yang tersedia:
//
//	?employee_id=1   : hanya cuti karyawan tertentu
//	?status=PENDING  : hanya cuti dengan status tertentu
func GetLeaves(c *gin.Context) {
	var leaves []models.Leave

	query := config.DB.Model(&models.Leave{}).Preload("Employee")

	employeeID, hasEmployee, ok := parseUintQuery(c, "employee_id")
	if !ok {
		return
	}
	if hasEmployee {
		query = query.Where("employee_id = ?", employeeID)
	}

	if status := strings.ToUpper(strings.TrimSpace(c.Query("status"))); status != "" {
		if !isValidLeaveStatus(status) {
			utils.Error(c, http.StatusBadRequest,
				"Query 'status' hanya boleh bernilai PENDING, APPROVED, atau REJECTED")
			return
		}
		query = query.Where("status = ?", status)
	}

	if err := query.Order("id DESC").Find(&leaves).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data cuti")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, "Daftar pengajuan cuti berhasil diambil",
		leaves, gin.H{"total": len(leaves)})
}

// ApproveLeave menangani PATCH /api/leaves/:id/approve, yaitu keputusan HR atas
// sebuah pengajuan cuti: APPROVED atau REJECTED.
//
// Bila disetujui, saldo jatah cuti karyawan langsung dihitung ulang. Seluruh
// proses dibungkus transaksi agar perubahan status cuti dan pembaruan saldo
// selalu berhasil bersama-sama, atau gagal bersama-sama.
func ApproveLeave(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var leave models.Leave
	if err := config.DB.First(&leave, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Pengajuan cuti tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data cuti")
		return
	}

	var req models.LeaveApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// Keputusan hanya boleh diambil satu kali atas pengajuan yang masih PENDING.
	if leave.Status != models.LeaveStatusPending {
		utils.Error(c, http.StatusConflict,
			"Pengajuan cuti ini sudah diproses sebelumnya dengan status "+leave.Status)
		return
	}

	var employee models.Employee
	if err := config.DB.First(&employee, leave.EmployeeID).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan pengaju")
		return
	}

	// Persetujuan ditolak bila sisa jatah cuti karyawan tidak mencukupi.
	if req.Status == models.LeaveStatusApproved && leave.TotalDays > employee.LeaveBalance {
		utils.Error(c, http.StatusUnprocessableEntity, fmt.Sprintf(
			"Sisa jatah cuti karyawan tidak mencukupi (sisa %d hari, diajukan %d hari)",
			employee.LeaveBalance, leave.TotalDays))
		return
	}

	// Tahun acuan saldo diambil dari tanggal mulai cuti.
	year := utils.YearOfPeriod(leave.StartDate)

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		leave.Status = req.Status
		if err := tx.Save(&leave).Error; err != nil {
			return err
		}
		// Saldo dihitung ulang dari nol berdasarkan seluruh cuti APPROVED tahun
		// berjalan, bukan sekadar dikurangi. Cara ini membuat hasilnya tetap
		// benar walaupun endpoint dipanggil berkali-kali (idempoten).
		return recalculateLeaveBalance(tx, leave.EmployeeID, year)
	})
	if err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal memproses persetujuan cuti", err.Error())
		return
	}

	config.DB.Preload("Employee").First(&leave, leave.ID)

	message := "Pengajuan cuti berhasil disetujui"
	if req.Status == models.LeaveStatusRejected {
		message = "Pengajuan cuti berhasil ditolak"
	}

	utils.Success(c, http.StatusOK, message, leave)
}

// ============================================================================
// LOGIKA SALDO CUTI
// ============================================================================

// recalculateLeaveBalance menghitung ulang sisa jatah cuti seorang karyawan
// untuk satu tahun tertentu, lalu menyimpannya ke tabel employees.
//
// Rumus: LeaveBalance = LeaveQuota - total hari cuti berstatus APPROVED
// yang tanggal mulainya berada pada tahun tersebut. Hasil negatif dinolkan.
//
// Fungsi ini dipakai bersama oleh proses persetujuan cuti dan proses payroll,
// sehingga kedua alur selalu memakai definisi saldo yang sama persis.
func recalculateLeaveBalance(tx *gorm.DB, employeeID uint, year string) error {
	var employee models.Employee
	if err := tx.First(&employee, employeeID).Error; err != nil {
		return err
	}

	// COALESCE dipakai agar karyawan yang belum pernah cuti menghasilkan 0,
	// bukan NULL yang akan gagal dipindai ke tipe int.
	var usedDays int
	err := tx.Model(&models.Leave{}).
		Select("COALESCE(SUM(total_days), 0)").
		Where("employee_id = ? AND status = ?", employeeID, models.LeaveStatusApproved).
		Where("start_date LIKE ?", year+"%").
		Scan(&usedDays).Error
	if err != nil {
		return err
	}

	balance := employee.LeaveQuota - usedDays
	if balance < 0 {
		balance = 0
	}

	return tx.Model(&models.Employee{}).
		Where("id = ?", employeeID).
		Update("leave_balance", balance).Error
}

// isValidLeaveStatus memeriksa apakah status cuti termasuk nilai yang
// diizinkan sistem.
func isValidLeaveStatus(status string) bool {
	switch status {
	case models.LeaveStatusPending,
		models.LeaveStatusApproved,
		models.LeaveStatusRejected:
		return true
	default:
		return false
	}
}
