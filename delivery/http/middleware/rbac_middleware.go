package middleware

import (
	"net/http"

	"go-hris-payroll-system/dto"

	"github.com/gin-gonic/gin"
)

// RequireRole adalah middleware RBAC yang memvalidasi role pengguna dari Gin Context.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// FLOW 1: Ambil role dari context yang disimpan oleh AuthMiddleware
		roleVal, exists := c.Get("role")
		if !exists {
			dto.RespondError(c, http.StatusUnauthorized, "Identitas peran (role) tidak ditemukan di konteks request", nil)
			c.Abort()
			return
		}

		userRole, ok := roleVal.(string)
		if !ok {
			dto.RespondError(c, http.StatusInternalServerError, "Tipe data role tidak valid", nil)
			c.Abort()
			return
		}

		// FLOW 2: Periksa apakah role user termasuk dalam daftar allowedRoles
		hasPermission := false
		for _, role := range allowedRoles {
			if userRole == role {
				hasPermission = true
				break
			}
		}

		// FLOW 3: Jika role tidak diizinkan, kembalikan HTTP status 403 Forbidden
		if !hasPermission {
			dto.RespondError(c, http.StatusForbidden, "Akses ditolak: Peran Anda tidak memiliki izin untuk fitur ini", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
