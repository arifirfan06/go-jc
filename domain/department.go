package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Department merepresentasikan entitas Departemen dan Anggaran Divisi.
type Department struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Budget    float64   `gorm:"type:numeric(15,2);not null;default:0" json:"budget"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName menentukan nama tabel spesifik sesuai soal `department_budgets`.
func (Department) TableName() string {
	return "department_budgets"
}

// DepartmentRepository mendefinisikan kontrak interface akses database untuk Departemen & Budget.
type DepartmentRepository interface {
	FindByID(ctx context.Context, id uint) (*Department, error)
	// FindByIDWithLock mengambil data budget departemen dengan Row-Level Locking (FOR UPDATE) dalam transaksi DB
	FindByIDWithLock(ctx context.Context, tx *gorm.DB, id uint) (*Department, error)
	// DeductBudget mengurangi anggaran departemen dalam transaksi DB
	DeductBudget(ctx context.Context, tx *gorm.DB, id uint, amount float64) error
}
