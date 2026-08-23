package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-jc-challange2/config"
	"go-jc-challange2/models"
	"go-jc-challange2/utils"
)

// ============================================================================
// CONTROLLER KARYAWAN - MANAJEMEN DATA RELASIONAL
// ============================================================================

// CreateEmployee menangani POST /api/employees.
//
// Selain validasi format lewat tag binding, controller ini WAJIB memastikan
// DepartmentID dan PositionID yang dikirim benar-benar ada di database.
// Tanpa pengecekan ini, request bisa menyisipkan Foreign Key yang menggantung.
func CreateEmployee(c *gin.Context) {
	var req models.EmployeeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// --- Validasi relasi: departemen harus ada ------------------------------
	var department models.Department
	if err := config.DB.First(&department, req.DepartmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusBadRequest,
				"Departemen dengan ID tersebut tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data departemen")
		return
	}

	// --- Validasi relasi: jabatan harus ada ---------------------------------
	var position models.Position
	if err := config.DB.First(&position, req.PositionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusBadRequest,
				"Jabatan dengan ID tersebut tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data jabatan")
		return
	}

	nik := strings.TrimSpace(req.NIK)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	// --- Validasi keunikan NIK dan Email ------------------------------------
	var duplicate models.Employee
	err := config.DB.Where("nik = ? OR email = ?", nik, email).First(&duplicate).Error
	if err == nil {
		field := "NIK"
		if duplicate.Email == email {
			field = "Email"
		}
		utils.Error(c, http.StatusConflict, field+" tersebut sudah terdaftar pada karyawan lain")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi karyawan")
		return
	}

	// Status boleh dikosongkan klien; karyawan baru dianggap ACTIVE.
	status := req.Status
	if status == "" {
		status = models.EmployeeStatusActive
	}

	// Jatah cuti memakai nilai bawaan bila tidak dikirim klien.
	leaveQuota := models.DefaultLeaveQuota
	if req.LeaveQuota != nil {
		leaveQuota = *req.LeaveQuota
	}

	employee := models.Employee{
		NIK:          nik,
		FullName:     strings.TrimSpace(req.FullName),
		Email:        email,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		Status:       status,
		LeaveQuota:   leaveQuota,
		// Karyawan baru belum memakai cuti, jadi saldo awal sama dengan jatah.
		LeaveBalance: leaveQuota,
	}

	if err := config.DB.Create(&employee).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan data karyawan", err.Error())
		return
	}

	// Muat ulang beserta relasinya supaya response langsung menampilkan
	// detail departemen dan jabatan karyawan yang baru dibuat.
	config.DB.Preload("Department").Preload("Position").First(&employee, employee.ID)

	utils.Success(c, http.StatusCreated, "Karyawan berhasil ditambahkan", employee)
}

// GetEmployees menangani GET /api/employees.
//
// Menampilkan seluruh karyawan LENGKAP dengan detail departemen dan jabatannya
// menggunakan Preload GORM (setara dengan SQL JOIN), serta mendukung:
//
//	?search=nama       : pencarian pada nama lengkap, NIK, atau email
//	?department_id=1   : filter berdasarkan departemen
//	?position_id=2     : filter berdasarkan jabatan
//	?status=ACTIVE     : filter berdasarkan status karyawan
//
// Seluruh filter di atas dapat digabungkan dalam satu request.
func GetEmployees(c *gin.Context) {
	var employees []models.Employee

	// Preload menerbitkan query tambahan untuk mengisi field Department dan
	// Position pada setiap karyawan, sehingga response berisi objek relasi utuh.
	query := config.DB.Model(&models.Employee{}).
		Preload("Department").
		Preload("Position")

	// --- Filter pencarian nama karyawan -------------------------------------
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		// LOWER(...) LIKE ... dipakai supaya pencarian tidak membedakan
		// huruf besar dan huruf kecil.
		keyword := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(full_name) LIKE ? OR LOWER(nik) LIKE ? OR LOWER(email) LIKE ?",
			keyword, keyword, keyword,
		)
	}

	// --- Filter departemen ---------------------------------------------------
	departmentID, hasDepartment, ok := parseUintQuery(c, "department_id")
	if !ok {
		return
	}
	if hasDepartment {
		query = query.Where("department_id = ?", departmentID)
	}

	// --- Filter jabatan ------------------------------------------------------
	positionID, hasPosition, ok := parseUintQuery(c, "position_id")
	if !ok {
		return
	}
	if hasPosition {
		query = query.Where("position_id = ?", positionID)
	}

	// --- Filter status karyawan ----------------------------------------------
	if status := strings.ToUpper(strings.TrimSpace(c.Query("status"))); status != "" {
		if !isValidEmployeeStatus(status) {
			utils.Error(c, http.StatusBadRequest,
				"Query 'status' hanya boleh bernilai ACTIVE, SUSPENDED, atau TERMINATED")
			return
		}
		query = query.Where("status = ?", status)
	}

	if err := query.Order("id ASC").Find(&employees).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, "Daftar karyawan berhasil diambil", employees, gin.H{
		"total": len(employees),
		"filters": gin.H{
			"search":        c.Query("search"),
			"department_id": c.Query("department_id"),
			"position_id":   c.Query("position_id"),
			"status":        c.Query("status"),
		},
	})
}

// GetEmployeeByID menangani GET /api/employees/:id.
func GetEmployeeByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var employee models.Employee
	err := config.DB.Preload("Department").Preload("Position").First(&employee, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Karyawan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan")
		return
	}

	utils.Success(c, http.StatusOK, "Detail karyawan berhasil diambil", employee)
}

// UpdateEmployee menangani PUT /api/employees/:id.
//
// Bersifat partial update: klien cukup mengirim field yang ingin diubah,
// termasuk department_id atau position_id untuk memindahkan karyawan ke
// departemen/jabatan lain. Setiap perpindahan tetap divalidasi keberadaannya.
func UpdateEmployee(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var employee models.Employee
	if err := config.DB.First(&employee, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Karyawan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan")
		return
	}

	var req models.EmployeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// --- Perpindahan departemen ----------------------------------------------
	if req.DepartmentID != nil {
		var department models.Department
		if err := config.DB.First(&department, *req.DepartmentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusBadRequest,
					"Departemen tujuan dengan ID tersebut tidak ditemukan")
				return
			}
			utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data departemen")
			return
		}
		employee.DepartmentID = *req.DepartmentID
	}

	// --- Perpindahan jabatan --------------------------------------------------
	if req.PositionID != nil {
		var position models.Position
		if err := config.DB.First(&position, *req.PositionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.Error(c, http.StatusBadRequest,
					"Jabatan tujuan dengan ID tersebut tidak ditemukan")
				return
			}
			utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data jabatan")
			return
		}
		employee.PositionID = *req.PositionID
	}

	// --- Perubahan NIK (tetap harus unik) -------------------------------------
	if req.NIK != nil {
		nik := strings.TrimSpace(*req.NIK)
		var duplicate models.Employee
		err := config.DB.Where("nik = ? AND id <> ?", nik, employee.ID).First(&duplicate).Error
		if err == nil {
			utils.Error(c, http.StatusConflict, "NIK tersebut sudah dipakai karyawan lain")
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi NIK")
			return
		}
		employee.NIK = nik
	}

	// --- Perubahan Email (tetap harus unik) -----------------------------------
	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		var duplicate models.Employee
		err := config.DB.Where("email = ? AND id <> ?", email, employee.ID).First(&duplicate).Error
		if err == nil {
			utils.Error(c, http.StatusConflict, "Email tersebut sudah dipakai karyawan lain")
			return
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa duplikasi email")
			return
		}
		employee.Email = email
	}

	if req.FullName != nil {
		employee.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Status != nil {
		employee.Status = *req.Status
	}
	if req.LeaveQuota != nil {
		// Saldo cuti ikut digeser sebesar selisih perubahan jatah, agar cuti yang
		// sudah terpakai tidak tiba-tiba hangus maupun terhitung dua kali.
		used := employee.LeaveQuota - employee.LeaveBalance
		employee.LeaveQuota = *req.LeaveQuota
		employee.LeaveBalance = employee.LeaveQuota - used
		if employee.LeaveBalance < 0 {
			employee.LeaveBalance = 0
		}
	}

	if err := config.DB.Save(&employee).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal memperbarui data karyawan", err.Error())
		return
	}

	config.DB.Preload("Department").Preload("Position").First(&employee, employee.ID)

	utils.Success(c, http.StatusOK, "Data karyawan berhasil diperbarui", employee)
}

// DeleteEmployee menangani DELETE /api/employees/:id.
//
// Data absensi, cuti, dan slip gaji milik karyawan ikut terhapus karena relasi
// tersebut didefinisikan dengan constraint OnDelete:CASCADE pada model.
func DeleteEmployee(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var employee models.Employee
	if err := config.DB.First(&employee, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Karyawan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan")
		return
	}

	if err := config.DB.Delete(&employee).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menghapus karyawan", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Karyawan berhasil dihapus", gin.H{"id": employee.ID})
}

// isValidEmployeeStatus memeriksa apakah string status termasuk salah satu dari
// tiga nilai yang diizinkan sistem.
func isValidEmployeeStatus(status string) bool {
	switch status {
	case models.EmployeeStatusActive,
		models.EmployeeStatusSuspended,
		models.EmployeeStatusTerminated:
		return true
	default:
		return false
	}
}
