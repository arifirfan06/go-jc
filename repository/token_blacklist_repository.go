package repository

import (
	"context"
	"errors"
	"go-hris-payroll-system/domain"
	"time"

	"gorm.io/gorm"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository menginisialisasi repository token blacklist.
func NewTokenBlacklistRepository(db *gorm.DB) domain.TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

// BlacklistToken memasukkan token JWT ke dalam daftar hitam di database.
func (r *tokenBlacklistRepository) BlacklistToken(ctx context.Context, token string, expiredAt time.Time) error {
	tb := &domain.TokenBlacklist{
		Token:     token,
		ExpiredAt: expiredAt,
	}
	// FLOW: Menjalankan kueri INSERT INTO token_blacklists (token, expired_at)
	return r.db.WithContext(ctx).Create(tb).Error
}

// IsBlacklisted mengecek apakah token sudah terdaftar di daftar hitam.
func (r *tokenBlacklistRepository) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	var count int64
	// FLOW: Menjalankan kueri SELECT count(*) FROM token_blacklists WHERE token = ?
	err := r.db.WithContext(ctx).Model(&domain.TokenBlacklist{}).Where("token = ?", token).Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return count > 0, nil
}
