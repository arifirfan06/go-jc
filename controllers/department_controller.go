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
// CONTROLLER DEPARTEMEN - CRUD DATA MASTER
// ============================================================================

// CreateDepartment menangani POST /api/departments.
// Membuat satu departemen baru setelah memastikan kodenya belum terpakai.
func CreateDepartment(c *gin.Context) {
	var req models.DepartmentRequest

	// ShouldBindJSON menjalankan seluruh tag `binding` pada DepartmentRequest.
	// Bila ada yang gagal, helper akan membalas 400 Bad Request.
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	code := strings.TrimSpace(req.Code)

	// Cek keunikan kode departemen sebelum menyimpan, agar pesan error yang
	// diterima klien lebih ramah dibanding error constraint mentah dari SQLite.
	var existing models.Department
	err := config.DB.Where("code = ?", code).First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusConflict,
			"Kode departemen '"+code+"' sudah digunakan oleh departemen lain")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa kode departemen")
		return
	}

	department := models.Department{
		Name: strings.TrimSpace(req.Name),
		Code: code,
	}

	if err := config.DB.Create(&department).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan departemen", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Departemen berhasil dibuat", department)
}

// GetDepartments menangani GET /api/departments.
// Mendukung pencarian bebas lewat ?search= yang mencocokkan nama maupun kode.
func GetDepartments(c *gin.Context) {
	var departments []models.Department

	query := config.DB.Model(&models.Department{})

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		// LOWER(...) LIKE ... dipakai supaya pencarian tidak membedakan
		// huruf besar dan huruf kecil.
		keyword := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", keyword, keyword)
	}

	if err := query.Order("id ASC").Find(&departments).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data departemen")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, "Daftar departemen berhasil diambil",
		departments, gin.H{"total": len(departments)})
}

// GetDepartmentByID menangani GET /api/departments/:id.
func GetDepartmentByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var department models.Department
	if err := config.DB.First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Departemen tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data departemen")
		return
	}

	utils.Success(c, http.StatusOK, "Detail departemen berhasil diambil", department)
}

// UpdateDepartment menangani PUT /api/departments/:id.
func UpdateDepartment(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var department models.Department
	if err := config.DB.First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Departemen tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data departemen")
		return
	}

	var req models.DepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	code := strings.TrimSpace(req.Code)

	// Pastikan kode baru tidak bentrok dengan departemen LAIN.
	// Klausa "id <> ?" membuat departemen ini boleh menyimpan kodenya sendiri.
	var duplicate models.Department
	err := config.DB.Where("code = ? AND id <> ?", code, department.ID).First(&duplicate).Error
	if err == nil {
		utils.Error(c, http.StatusConflict,
			"Kode departemen '"+code+"' sudah digunakan oleh departemen lain")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa kode departemen")
		return
	}

	department.Name = strings.TrimSpace(req.Name)
	department.Code = code

	if err := config.DB.Save(&department).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal memperbarui departemen", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Departemen berhasil diperbarui", department)
}

// DeleteDepartment menangani DELETE /api/departments/:id.
//
// Penghapusan ditolak bila departemen masih menaungi karyawan. Pengecekan ini
// dilakukan lebih dulu di level aplikasi supaya klien menerima pesan yang jelas,
// bukan sekadar error constraint Foreign Key dari SQLite.
func DeleteDepartment(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var department models.Department
	if err := config.DB.First(&department, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Departemen tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data departemen")
		return
	}

	var employeeCount int64
	if err := config.DB.Model(&models.Employee{}).
		Where("department_id = ?", department.ID).Count(&employeeCount).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa relasi karyawan")
		return
	}

	if employeeCount > 0 {
		utils.Error(c, http.StatusConflict,
			"Departemen tidak dapat dihapus karena masih dipakai oleh karyawan aktif di sistem")
		return
	}

	if err := config.DB.Delete(&department).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menghapus departemen", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Departemen berhasil dihapus", gin.H{"id": department.ID})
}
