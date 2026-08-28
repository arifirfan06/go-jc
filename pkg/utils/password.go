package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah plain text password menjadi hash Bcrypt.
func HashPassword(password string) (string, error) {
	// FLOW: Mengamankan password menggunakan Bcrypt dengan cost DefaultCost (10)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash membandingkan plain text password dengan hash Bcrypt.
func CheckPasswordHash(password, hash string) bool {
	// FLOW: Memverifikasi kesamaan password dengan hash yang tersimpan
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
