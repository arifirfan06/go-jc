package repository

import (
	"context"
	"errors"
	"go-hris-payroll-system/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository menginisialisasi repository user dengan instance GORM DB.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create menyimpan data user baru ke database.
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// FLOW: Menjalankan kueri INSERT INTO users
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByEmail mencari user berdasarkan email beserta relasi departemennya.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	// FLOW: Menjalankan kueri SELECT * FROM users WHERE email = ? LIMIT 1
	err := r.db.WithContext(ctx).Preload("Department").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByID mencari user berdasarkan ID.
func (r *userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	// FLOW: Menjalankan kueri SELECT * FROM users WHERE id = ? LIMIT 1
	err := r.db.WithContext(ctx).Preload("Department").First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
