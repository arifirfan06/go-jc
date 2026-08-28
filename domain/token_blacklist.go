package domain

import (
	"context"
	"time"
)

// TokenBlacklist merepresentasikan entitas Token JWT yang telah di-logout/dibatalkan.
type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"type:text;uniqueIndex;not null" json:"token"`
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName menentukan nama tabel `token_blacklists`.
func (TokenBlacklist) TableName() string {
	return "token_blacklists"
}

// TokenBlacklistRepository mendefinisikan kontrak interface untuk mengelola token blacklist di DB.
type TokenBlacklistRepository interface {
	BlacklistToken(ctx context.Context, token string, expiredAt time.Time) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}
