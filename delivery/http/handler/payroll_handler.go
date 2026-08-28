package handler

import (
	"errors"
	"net/http"

	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"
	"go-hris-payroll-system/usecase"

	"github.com/gin-gonic/gin"
)

type PayrollHandler struct {
	payrollUsecase domain.PayrollUsecase
}

// NewPayrollHandler menginisialisasi Handler HTTP Penggajian.
func NewPayrollHandler(payrollUsecase domain.PayrollUsecase) *PayrollHandler {
	return &PayrollHandler{payrollUsecase: payrollUsecase}
}

// ProcessPayroll (POST /api/v1/payroll/process) - Hanya dapat diakses oleh HRD.
func (h *PayrollHandler) ProcessPayroll(c *gin.Context) {
	var req dto.ProcessPayrollRequest

	// FLOW 1: Bind JSON Request Body list EmployeeID & Bonus
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondValidationError(c, err)
		return
	}

	if len(req.Payrolls) == 0 {
		dto.RespondError(c, http.StatusBadRequest, "Daftar penggajian (payrolls) tidak boleh kosong", nil)
		return
	}

	// FLOW 2: Map DTO request ke domain PayrollItemRequest
	var items []domain.PayrollItemRequest
	for _, p := range req.Payrolls {
		items = append(items, domain.PayrollItemRequest{
			EmployeeID: p.EmployeeID,
			Bonus:      p.Bonus,
		})
	}

	// FLOW 3: Panggil Usecase ProcessPayroll (Menjalankan DB Transaction & Row Locking)
	payrolls, err := h.payrollUsecase.ProcessPayroll(c.Request.Context(), items)
	if err != nil {
		// FLOW 4: Jika sisa budget departemen tidak mencukupi atau terjadi error lain,
		// kembalikan error informatif ke client (Seluruh transaksi sudah otomatis di-rollback)
		if errors.Is(err, domain.ErrInsufficientBudget) {
			dto.RespondError(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		dto.RespondError(c, http.StatusInternalServerError, "Gagal memproses penggajian bulanan", err.Error())
		return
	}

	// FLOW 5: Format response DTO dan kembalikan status 200 OK
	response := usecase.ConvertToPayrollResponses(payrolls)
	dto.RespondSuccess(c, http.StatusOK, "Penggajian bulanan berhasil diproses dan anggaran departemen telah dipotong", response)
}
