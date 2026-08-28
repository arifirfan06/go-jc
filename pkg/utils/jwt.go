package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims mendefinisikan payload data yang disimpan di dalam JWT token.
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT Token membuat token JWT baru dengan masa berlaku tertentu (misal 24 jam).
func GenerateJWT(userID uint, email, role, secretKey string) (string, error) {
	// FLOW 1: Menentukan klaim JWT (user_id, email, role, dan waktu kedaluwarsa)
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// FLOW 2: Membuat token dengan algoritma pembuat tanda tangan HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// FLOW 3: Membubuhi tanda tangan pada token menggunakan secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ParseJWT memvalidasi token string dan mengembalikan data claims jika valid.
func ParseJWT(tokenString, secretKey string) (*JWTClaims, error) {
	// FLOW 1: Melakukan parsing token string dan memverifikasi algoritma serta secret key
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritma signing token tidak valid")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// FLOW 2: Mengambil klaim data jika token valid
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid atau telah kedaluwarsa")
	}

	return claims, nil
}
