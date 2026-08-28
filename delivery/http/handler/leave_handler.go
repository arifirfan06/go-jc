package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"
	"go-hris-payroll-system/usecase"

	"github.com/gin-gonic/gin"
)

type LeaveHandler struct {
	leaveUsecase domain.LeaveUsecase
}

// NewLeaveHandler menginisialisasi Handler HTTP Pengajuan Cuti.
func NewLeaveHandler(leaveUsecase domain.LeaveUsecase) *LeaveHandler {
	return &LeaveHandler{leaveUsecase: leaveUsecase}
}

// CreateLeave (POST /api/v1/leaves) - Dapat diakses oleh Employee & HRD.
func (h *LeaveHandler) CreateLeave(c *gin.Context) {
	var req dto.CreateLeaveRequest

	// FLOW 1: Bind JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondValidationError(c, err)
		return
	}

	// FLOW 2: Ambil UserID dari Gin Context (diisi oleh AuthMiddleware)
	userIDVal, _ := c.Get("userID")
	employeeID := userIDVal.(uint)

	// FLOW 3: Parse format tanggal start_date dan end_date (YYYY-MM-DD)
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		dto.RespondError(c, http.StatusBadRequest, "Format start_date harus YYYY-MM-DD", err.Error())
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		dto.RespondError(c, http.StatusBadRequest, "Format end_date harus YYYY-MM-DD", err.Error())
		return
	}

	// FLOW 4: Panggil Usecase CreateLeave
	leave, err := h.leaveUsecase.CreateLeave(c.Request.Context(), employeeID, req.Reason, startDate, endDate)
	if err != nil {
		dto.RespondError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// FLOW 5: Kembalikan response sukses 201 Created
	res := usecase.ConvertToLeaveResponse(leave)
	dto.RespondSuccess(c, http.StatusCreated, "Pengajuan cuti berhasil dibuat", res)
}

// GetLeaveByID (GET /api/v1/leaves/:id) - Dengan Mitigasi IDOR Defensif.
func (h *LeaveHandler) GetLeaveByID(c *gin.Context) {
	// FLOW 1: Ambil parameter ID cuti dari URL path
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		dto.RespondError(c, http.StatusBadRequest, "ID cuti harus berupa angka", nil)
		return
	}

	// FLOW 2: Ambil identitas pengguna dari Gin Context (ID & Role)
	userIDVal, _ := c.Get("userID")
	roleVal, _ := c.Get("role")
	requesterID := userIDVal.(uint)
	requesterRole := roleVal.(string)

	// FLOW 3: Panggil Usecase dengan pemeriksaan IDOR defensif
	leave, err := h.leaveUsecase.GetLeaveByID(c.Request.Context(), uint(id), requesterID, requesterRole)
	if err != nil {
		// FLOW 4: Jika karyawan biasa mencoba melihat cuti orang lain, kueri defensif mengembalikan ErrNotFound
		// Ini mencegah terornya celah keamanan IDOR
		if errors.Is(err, domain.ErrNotFound) {
			dto.RespondError(c, http.StatusNotFound, "Data pengajuan cuti tidak ditemukan atau Anda tidak memiliki akses", nil)
			return
		}
		dto.RespondError(c, http.StatusInternalServerError, "Gagal mengambil data cuti", err.Error())
		return
	}

	// FLOW 5: Kembalikan detail data cuti
	res := usecase.ConvertToLeaveResponse(leave)
	dto.RespondSuccess(c, http.StatusOK, "Berhasil mengambil detail pengajuan cuti", res)
}

// GetLeaves (GET /api/v1/leaves) - Mengambil daftar seluruh cuti (HRD) atau cuti pribadi (EMPLOYEE).
func (h *LeaveHandler) GetLeaves(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	roleVal, _ := c.Get("role")
	requesterID := userIDVal.(uint)
	requesterRole := roleVal.(string)

	leaves, err := h.leaveUsecase.GetLeaves(c.Request.Context(), requesterID, requesterRole)
	if err != nil {
		dto.RespondError(c, http.StatusInternalServerError, "Gagal mengambil daftar cuti", err.Error())
		return
	}

	res := usecase.ConvertToLeaveResponses(leaves)
	dto.RespondSuccess(c, http.StatusOK, "Berhasil mengambil daftar pengajuan cuti", res)
}
