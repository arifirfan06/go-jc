package handler

import (
	"errors"
	"net/http"

	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase domain.AuthUsecase
}

// NewAuthHandler menginisialisasi HTTP Handler untuk Autentikasi.
func NewAuthHandler(authUsecase domain.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Register (POST /api/v1/auth/register) - Hanya dapat diakses oleh HRD.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// FLOW 1: Bind JSON payload dan jalankan validasi (termasuk custom validator @company.co.id)
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondValidationError(c, err)
		return
	}

	// FLOW 2: Panggil usecase Register
	user, token, err := h.authUsecase.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.Role, req.DepartmentID, req.BasicSalary)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidEmailDomain) {
			dto.RespondError(c, http.StatusBadRequest, err.Error(), nil)
			return
		}
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			dto.RespondError(c, http.StatusConflict, err.Error(), nil)
			return
		}
		dto.RespondError(c, http.StatusInternalServerError, "Gagal mendaftarkan karyawan", err.Error())
		return
	}

	// FLOW 3: Format balasan DTO
	res := dto.AuthResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		Token: token,
	}

	// FLOW 4: Kembalikan response sukses 201 Created
	dto.RespondSuccess(c, http.StatusCreated, "Registrasi akun karyawan berhasil", res)
}

// Login (POST /api/v1/auth/login) - Dapat diakses oleh HRD dan Employee.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// FLOW 1: Bind JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.RespondValidationError(c, err)
		return
	}

	// FLOW 2: Panggil usecase Login
	user, token, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			dto.RespondError(c, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		dto.RespondError(c, http.StatusInternalServerError, "Gagal melakukan proses login", err.Error())
		return
	}

	// FLOW 3: Format balasan DTO
	res := dto.AuthResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		Token: token,
	}

	// FLOW 4: Kembalikan response sukses 200 OK beserta Token JWT
	dto.RespondSuccess(c, http.StatusOK, "Login berhasil", res)
}

// Logout (POST /api/v1/auth/logout) - Memasukkan token JWT ke dalam blacklist database.
func (h *AuthHandler) Logout(c *gin.Context) {
	// FLOW 1: Ambil string token dari Gin Context yang sudah di-set oleh AuthMiddleware
	tokenVal, exists := c.Get("token")
	if !exists {
		dto.RespondError(c, http.StatusBadRequest, "Token tidak ditemukan di konteks request", nil)
		return
	}

	tokenString := tokenVal.(string)

	// FLOW 2: Panggil usecase Logout untuk menyimpan token ke DB blacklist
	if err := h.authUsecase.Logout(c.Request.Context(), tokenString); err != nil {
		dto.RespondError(c, http.StatusInternalServerError, "Gagal melakukan logout", err.Error())
		return
	}

	// FLOW 3: Kembalikan response sukses
	dto.RespondSuccess(c, http.StatusOK, "Logout berhasil. Token telah dibatalkan.", nil)
}
