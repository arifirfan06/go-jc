package utils_test

import (
	"testing"

	"go-hris-payroll-system/pkg/utils"
)

func TestJWTAndHash(t *testing.T) {
	// Test Password Hashing
	password := "securepassword123"
	hash, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !utils.CheckPasswordHash(password, hash) {
		t.Errorf("password hash verification failed")
	}
	if utils.CheckPasswordHash("wrongpassword", hash) {
		t.Errorf("expected false for incorrect password")
	}

	// Test JWT
	secret := "my-secret-key"
	token, err := utils.GenerateJWT(1, "user@company.co.id", "EMPLOYEE", secret)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := utils.ParseJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	if claims.UserID != 1 || claims.Email != "user@company.co.id" || claims.Role != "EMPLOYEE" {
		t.Errorf("claims mismatch: got %+v", claims)
	}
}
