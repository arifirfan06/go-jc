package http

import (
	"log"

	"go-hris-payroll-system/delivery/http/handler"
	"go-hris-payroll-system/delivery/http/middleware"
	"go-hris-payroll-system/domain"
	customValidator "go-hris-payroll-system/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RouterConfig menyimpan dependency handler & middleware untuk mengonfigurasi router Gin.
type RouterConfig struct {
	Engine        *gin.Engine
	JWTSecret     string
	AuthHandler   *handler.AuthHandler
	PayrollHandler *handler.PayrollHandler
	LeaveHandler   *handler.LeaveHandler
	BlacklistRepo domain.TokenBlacklistRepository
}

// SetupRouter mendaftarkan seluruh endpoint API, middleware JWT, RBAC, dan custom validator.
func SetupRouter(cfg RouterConfig) {
	// FLOW 1: Registrasi Custom Validator (company_email) pada Gin Binding Engine
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("company_email", customValidator.CompanyEmailValidator); err != nil {
			log.Fatalf("[ERROR] Gagal mendaftarkan custom validator company_email: %v", err)
		}
		log.Println("[SUCCESS] Berhasil mendaftarkan custom validator '@company.co.id'")
	}

	// FLOW 2: Membuat Group API /api/v1
	v1 := cfg.Engine.Group("/api/v1")

	// ----------------------------------------------------
	// ROUTE AUTENTIKASI PUBLIC
	// ----------------------------------------------------
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/login", cfg.AuthHandler.Login)
	}

	// ----------------------------------------------------
	// ROUTE TERPROTEKSI JWT AUTHENTICATION
	// ----------------------------------------------------
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, cfg.BlacklistRepo))
	{
		// Logout (Dapat diakses oleh siapa saja yang terautentikasi)
		protected.POST("/auth/logout", cfg.AuthHandler.Logout)

		// ----------------------------------------------------
		// KHUSUS ROLE HRD (RBAC RequireRole("HRD"))
		// ----------------------------------------------------
		hrdOnly := protected.Group("")
		hrdOnly.Use(middleware.RequireRole(domain.RoleHRD))
		{
			// Registrasi akun karyawan baru oleh HRD
			hrdOnly.POST("/auth/register", cfg.AuthHandler.Register)

			// Proses penggajian bulanan & pemotongan anggaran departemen (DB Transaction + Row Lock)
			hrdOnly.POST("/payroll/process", cfg.PayrollHandler.ProcessPayroll)
		}

		// ----------------------------------------------------
		// FITUR CUTI (EMPLOYEE & HRD)
		// ----------------------------------------------------
		leaveRoutes := protected.Group("/leaves")
		leaveRoutes.Use(middleware.RequireRole(domain.RoleEmployee, domain.RoleHRD))
		{
			// Pengajuan Cuti (POST /api/v1/leaves)
			leaveRoutes.POST("", cfg.LeaveHandler.CreateLeave)

			// Detail Cuti (GET /api/v1/leaves/:id) - Dengan Mitigasi IDOR Defensif
			leaveRoutes.GET("/:id", cfg.LeaveHandler.GetLeaveByID)

			// Daftar Cuti (GET /api/v1/leaves)
			leaveRoutes.GET("", cfg.LeaveHandler.GetLeaves)
		}
	}
}
