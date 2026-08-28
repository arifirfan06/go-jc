package middleware

import (
	"net/http"
	"strings"

	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"
	"go-hris-payroll-system/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware mengamankan endpoint sensitif menggunakan JWT Token dan pemeriksaan Token Blacklist DB.
func AuthMiddleware(jwtSecret string, blacklistRepo domain.TokenBlacklistRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// FLOW 1: Ambil header Authorization dari request HTTP
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			dto.RespondError(c, http.StatusUnauthorized, "Header Authorization tidak ditemukan", nil)
			c.Abort()
			return
		}

		// FLOW 2: Pastikan format token adalah "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			dto.RespondError(c, http.StatusUnauthorized, "Format token Authorization harus Bearer <token>", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// FLOW 3: Periksa apakah token berada di Token Blacklist (Database)
		// Jika user sudah pernah logout, token tidak dapat digunakan kembali
		isBlacklisted, err := blacklistRepo.IsBlacklisted(c.Request.Context(), tokenString)
		if err != nil {
			dto.RespondError(c, http.StatusInternalServerError, "Gagal memverifikasi status token", err.Error())
			c.Abort()
			return
		}
		if isBlacklisted {
			dto.RespondError(c, http.StatusUnauthorized, "Token telah di-logout/dibatalkan. Silakan login kembali.", nil)
			c.Abort()
			return
		}

		// FLOW 4: Parse dan verifikasi tanda tangan serta klaim JWT
		claims, err := utils.ParseJWT(tokenString, jwtSecret)
		if err != nil {
			dto.RespondError(c, http.StatusUnauthorized, "Token JWT tidak valid atau telah expired", err.Error())
			c.Abort()
			return
		}

		// FLOW 5: Simpan informasi pengguna ke dalam Gin Context untuk digunakan di Handler berikutnya
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("token", tokenString)

		c.Next()
	}
}
