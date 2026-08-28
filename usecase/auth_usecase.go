package usecase

import (
	"context"
	"errors"
	"fmt"
	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/pkg/utils"
	"strings"
	"time"
)

type authUsecase struct {
	userRepo           domain.UserRepository
	tokenBlacklistRepo domain.TokenBlacklistRepository
	jwtSecret          string
}

// NewAuthUsecase menginisialisasi usecase autentikasi.
func NewAuthUsecase(userRepo domain.UserRepository, tokenBlacklistRepo domain.TokenBlacklistRepository, jwtSecret string) domain.AuthUsecase {
	return &authUsecase{
		userRepo:           userRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
		jwtSecret:          jwtSecret,
	}
}

// Register memproses pendaftaran karyawan baru oleh HRD.
func (u *authUsecase) Register(ctx context.Context, name, email, password, role string, departmentID uint, basicSalary float64) (*domain.User, string, error) {
	// FLOW 1: Validasi domain email harus diakhiri @company.co.id
	if !strings.HasSuffix(strings.ToLower(email), "@company.co.id") {
		return nil, "", domain.ErrInvalidEmailDomain
	}

	// FLOW 2: Cek apakah email sudah terdaftar di database
	existingUser, err := u.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", domain.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, "", fmt.Errorf("gagal mengecek email: %w", err)
	}

	// FLOW 3: Hash password menggunakan Bcrypt
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("gagal melakukan hash password: %w", err)
	}

	// FLOW 4: Buat objek User dan simpan ke database
	user := &domain.User{
		Name:         name,
		Email:        email,
		Password:     hashedPassword,
		Role:         role,
		DepartmentID: departmentID,
		BasicSalary:  basicSalary,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("gagal menyimpan data user: %w", err)
	}

	// FLOW 5: Generate JWT Token untuk user baru
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, u.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat token JWT: %w", err)
	}

	return user, token, nil
}

// Login memproses autentikasi pengguna dan mengembalikan JWT Token.
func (u *authUsecase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	// FLOW 1: Cari user berdasarkan email
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	// FLOW 2: Verifikasi kecocokan password Bcrypt
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", domain.ErrInvalidCredentials
	}

	// FLOW 3: Generate JWT Token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, u.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat token JWT: %w", err)
	}

	return user, token, nil
}

// Logout memasukkan JWT token aktif ke dalam daftar hitam database.
func (u *authUsecase) Logout(ctx context.Context, tokenString string) error {
	// FLOW 1: Parse token untuk mendapatkan waktu kedaluwarsa (expires_at)
	claims, err := utils.ParseJWT(tokenString, u.jwtSecret)
	var expiredAt time.Time
	if err == nil && claims.ExpiresAt != nil {
		expiredAt = claims.ExpiresAt.Time
	} else {
		expiredAt = time.Now().Add(24 * time.Hour)
	}

	// FLOW 2: Simpan token ke database blacklist
	if err := u.tokenBlacklistRepo.BlacklistToken(ctx, tokenString, expiredAt); err != nil {
		return fmt.Errorf("gagal memasukkan token ke blacklist: %w", err)
	}

	return nil
}
