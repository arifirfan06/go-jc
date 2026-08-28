package repository

import (
	"context"
	"errors"
	"fmt"
	"go-hris-payroll-system/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type departmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository menginisialisasi repository department.
func NewDepartmentRepository(db *gorm.DB) domain.DepartmentRepository {
	return &departmentRepository{db: db}
}

// FindByID mengambil data departemen tanpa locking.
func (r *departmentRepository) FindByID(ctx context.Context, id uint) (*domain.Department, error) {
	var dept domain.Department
	// FLOW: Menjalankan kueri SELECT * FROM department_budgets WHERE id = ?
	err := r.db.WithContext(ctx).First(&dept, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &dept, nil
}

// FindByIDWithLock mengambil data departemen dalam transaksi dengan Row-Level Locking (FOR UPDATE).
func (r *departmentRepository) FindByIDWithLock(ctx context.Context, tx *gorm.DB, id uint) (*domain.Department, error) {
	var dept domain.Department

	// FLOW 1: Gunakan koneksi transaksi (tx) jika ada, atau db default jika nil
	dbConn := r.db
	if tx != nil {
		dbConn = tx
	}

	// FLOW 2: Eksekusi SELECT * FROM department_budgets WHERE id = ? FOR UPDATE
	// Kunci baris (Row-Level Locking) untuk mencegah race condition saat pengubahan budget concurrent
	err := dbConn.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&dept, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &dept, nil
}

// DeductBudget mengurangi sisa anggaran departemen dalam transaksi.
func (r *departmentRepository) DeductBudget(ctx context.Context, tx *gorm.DB, id uint, amount float64) error {
	dbConn := r.db
	if tx != nil {
		dbConn = tx
	}

	// FLOW: Update sisa budget departemen: budget = budget - amount
	res := dbConn.WithContext(ctx).Model(&domain.Department{}).
		Where("id = ? AND budget >= ?", id, amount).
		Update("budget", gorm.Expr("budget - ?", amount))

	if res.Error != nil {
		return res.Error
	}

	// FLOW: Jika baris yang di-update 0, berarti budget tidak mencukupi
	if res.RowsAffected == 0 {
		return fmt.Errorf("%w: budget tidak mencukupi untuk pemotongan sejumlah %.2f", domain.ErrInsufficientBudget, amount)
	}

	return nil
}
