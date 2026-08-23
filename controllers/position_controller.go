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
// CONTROLLER JABATAN - CRUD DATA MASTER
// ============================================================================

// CreatePosition menangani POST /api/positions.
func CreatePosition(c *gin.Context) {
	var req models.PositionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	title := strings.TrimSpace(req.Title)

	// Nama jabatan dijaga tetap unik agar tidak muncul dua "Software Engineer"
	// dengan gaji pokok berbeda yang membingungkan saat payroll.
	var existing models.Position
	err := config.DB.Where("title = ?", title).First(&existing).Error
	if err == nil {
		utils.Error(c, http.StatusConflict, "Jabatan '"+title+"' sudah terdaftar")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data jabatan")
		return
	}

	position := models.Position{
		Title:      title,
		BaseSalary: req.BaseSalary,
	}

	if err := config.DB.Create(&position).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menyimpan jabatan", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Jabatan berhasil dibuat", position)
}

// GetPositions menangani GET /api/positions.
// Mendukung pencarian judul jabatan lewat ?search=.
func GetPositions(c *gin.Context) {
	var positions []models.Position

	query := config.DB.Model(&models.Position{})

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		keyword := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ?", keyword)
	}

	if err := query.Order("id ASC").Find(&positions).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data jabatan")
		return
	}

	utils.SuccessWithMeta(c, http.StatusOK, "Daftar jabatan berhasil diambil",
		positions, gin.H{"total": len(positions)})
}

// GetPositionByID menangani GET /api/positions/:id.
func GetPositionByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Jabatan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data jabatan")
		return
	}

	utils.Success(c, http.StatusOK, "Detail jabatan berhasil diambil", position)
}

// UpdatePosition menangani PUT /api/positions/:id.
//
// Catatan: mengubah BaseSalary di sini TIDAK mengubah slip gaji yang sudah
// terbit, karena tabel salaries menyimpan salinan (snapshot) gaji pokok pada
// saat payroll dijalankan.
func UpdatePosition(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Jabatan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data jabatan")
		return
	}

	var req models.PositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	title := strings.TrimSpace(req.Title)

	var duplicate models.Position
	err := config.DB.Where("title = ? AND id <> ?", title, position.ID).First(&duplicate).Error
	if err == nil {
		utils.Error(c, http.StatusConflict, "Jabatan '"+title+"' sudah terdaftar")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa data jabatan")
		return
	}

	position.Title = title
	position.BaseSalary = req.BaseSalary

	if err := config.DB.Save(&position).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal memperbarui jabatan", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Jabatan berhasil diperbarui", position)
}

// DeletePosition menangani DELETE /api/positions/:id.
// Sama seperti departemen, jabatan yang masih dipakai karyawan tidak boleh dihapus.
func DeletePosition(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var position models.Position
	if err := config.DB.First(&position, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Error(c, http.StatusNotFound, "Jabatan tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data jabatan")
		return
	}

	var employeeCount int64
	if err := config.DB.Model(&models.Employee{}).
		Where("position_id = ?", position.ID).Count(&employeeCount).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal memeriksa relasi karyawan")
		return
	}

	if employeeCount > 0 {
		utils.Error(c, http.StatusConflict,
			"Jabatan tidak dapat dihapus karena masih dipegang oleh karyawan di sistem")
		return
	}

	if err := config.DB.Delete(&position).Error; err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menghapus jabatan", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Jabatan berhasil dihapus", gin.H{"id": position.ID})
}
