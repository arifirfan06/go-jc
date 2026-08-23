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
// CONTROLLER KEHADIRAN (ABSENSI)
// ============================================================================

// CreateAttendance menangani POST /api/attendances, yaitu pencatatan jam masuk
// (check-in) karyawan.
//
// Aturan yang diterapkan:
//  1. Karyawan harus ada dan berstatus ACTIVE.
//  2. Satu karyawan hanya boleh punya satu catatan kehadiran per tanggal.
//  3. Bila Status tidak dikirim klien, sistem menentukannya sendiri dengan
//     membandingkan jam masuk terhadap WorkStartTime (08:00).
//  4. CheckOut boleh langsung diisi di sini, atau menyusul lewat endpoint
//     PATCH /api/attendances/:id/checkout.
func CreateAttendance(c *gin.Context) {
	var req models.AttendanceCheckInRequest

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

	// Karyawan non-aktif tidak berhak mencatatkan kehadiran.
	if employee.Status != models.EmployeeStatusActive {
		utils.Error(c, http.StatusUnprocessableEntity,
			"Karyawan berstatus "+employee.Status+" tidak dapat melakukan absensi")
		return
	}

	// --- Cegah check-in ganda pada tanggal yang sama ------------------------
	var existing models.Attendance
	err := config.DB.Where("employee_id = ? AND date = ?", req.EmployeeID, req.Date).
		First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusConflict,
			"Karyawan sudah memiliki catatan kehadiran pada tanggal "+req.Date)
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa catatan kehadiran")
		return
	}

	// --- Validasi urutan jam: pulang tidak boleh mendahului masuk -----------
	if req.CheckOut != "" && !utils.IsTimeAfter(req.CheckOut, req.CheckIn) {
		utils.Error(c, http.StatusBadRequest,
			"Jam pulang (check_out) harus lebih lambat daripada jam masuk (check_in)")
		return
	}

	// --- Penentuan status kehadiran ------------------------------------------
	status := req.Status
	if status == "" {
		status = determineAttendanceStatus(req.CheckIn)
	}

	attendance := models.Attendance{
		EmployeeID: req.EmployeeID,
		Date:       req.Date,
		CheckIn:    req.CheckIn,
		CheckOut:   req.CheckOut,
		Status:     status,
	}

	if err := config.DB.Create(&attendance).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan catatan kehadiran", err.Error())
		return
	}

	config.DB.Preload("Employee").First(&attendance, attendance.ID)

	utils.Success(c, http.StatusCreated, "Check-in berhasil dicatat", attendance)
}

// CheckOutAttendance menangani PATCH /api/attendances/:id/checkout,
// yaitu pencatatan jam pulang pada baris kehadiran yang sudah ada.
func CheckOutAttendance(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var attendance models.Attendance
	if err := config.DB.First(&attendance, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Catatan kehadiran tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil catatan kehadiran")
		return
	}

	var req models.AttendanceCheckOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// Check-out hanya masuk akal bila karyawan memang tercatat hadir.
	if attendance.Status == models.AttendanceStatusAbsent {
		utils.Error(c, http.StatusUnprocessableEntity,
			"Catatan berstatus ABSENT tidak dapat diberi jam pulang")
		return
	}

	if attendance.CheckOut != "" {
		utils.Error(c, http.StatusConflict,
			"Karyawan sudah melakukan check-out pada pukul "+attendance.CheckOut)
		return
	}

	if !utils.IsTimeAfter(req.CheckOut, attendance.CheckIn) {
		utils.Error(c, http.StatusBadRequest,
			"Jam pulang harus lebih lambat daripada jam masuk ("+attendance.CheckIn+")")
		return
	}

	attendance.CheckOut = req.CheckOut

	if err := config.DB.Save(&attendance).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan jam pulang", err.Error())
		return
	}

	config.DB.Preload("Employee").First(&attendance, attendance.ID)

	utils.Success(c, http.StatusOK, "Check-out berhasil dicatat", attendance)
}

// GetAttendances menangani GET /api/attendances.
//
// Filter yang tersedia:
//
//	?employee_id=1     : hanya kehadiran karyawan tertentu
//	?period=2026-08    : hanya kehadiran pada satu periode bulanan
//	?date=2026-08-17   : hanya kehadiran pada satu tanggal
//	?status=LATE       : hanya kehadiran dengan status tertentu
func GetAttendances(c *gin.Context) {
	var attendances []models.Attendance

	query := config.DB.Model(&models.Attendance{}).Preload("Employee")

	employeeID, hasEmployee, ok := parseUintQuery(c, "employee_id")
	if !ok {
		return
	}
	if hasEmployee {
		query = query.Where("employee_id = ?", employeeID)
	}

	if period := strings.TrimSpace(c.Query("period")); period != "" {
		startDate, endDate, err := utils.PeriodRange(period)
		if err != nil {
			utils.Error(c, http.StatusBadRequest,
				"Query 'period' harus mengikuti format YYYY-MM")
			return
		}
		// Tanggal disimpan sebagai string YYYY-MM-DD sehingga urutan leksikalnya
		// sama dengan urutan kronologisnya. Perbandingan BETWEEN aman dipakai.
		query = query.Where("date BETWEEN ? AND ?", startDate, endDate)
	}

	if date := strings.TrimSpace(c.Query("date")); date != "" {
		if _, err := utils.ParseDate(date); err != nil {
			utils.Error(c, http.StatusBadRequest,
				"Query 'date' harus mengikuti format YYYY-MM-DD")
			return
		}
		query = query.Where("date = ?", date)
	}

	if status := strings.ToUpper(strings.TrimSpace(c.Query("status"))); status != "" {
		if !isValidAttendanceStatus(status) {
			utils.Error(c, http.StatusBadRequest,
				"Query 'status' hanya boleh bernilai PRESENT, LATE, atau ABSENT")
			return
		}
		query = query.Where("status = ?", status)
	}

	if err := query.Order("date DESC, id DESC").Find(&attendances).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data kehadiran")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, "Daftar kehadiran berhasil diambil",
		attendances, gin.H{"total": len(attendances)})
}

// determineAttendanceStatus menentukan status kehadiran dari jam masuk.
// Check-in melewati WorkStartTime (08:00) dianggap terlambat.
func determineAttendanceStatus(checkIn string) string {
	if utils.IsTimeAfter(checkIn, models.WorkStartTime) {
		return models.AttendanceStatusLate
	}
	return models.AttendanceStatusPresent
}

// isValidAttendanceStatus memeriksa apakah status kehadiran termasuk nilai
// yang diizinkan sistem.
func isValidAttendanceStatus(status string) bool {
	switch status {
	case models.AttendanceStatusPresent,
		models.AttendanceStatusLate,
		models.AttendanceStatusAbsent:
		return true
	default:
		return false
	}
}
