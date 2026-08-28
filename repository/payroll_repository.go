package repository

import (
	"context"
	"go-hris-payroll-system/domain"

	"gorm.io/gorm"
)

type payrollRepository struct {
	db *gorm.DB
}

// NewPayrollRepository menginisialisasi repository payroll.
func NewPayrollRepository(db *gorm.DB) domain.PayrollRepository {
	return &payrollRepository{db: db}
}

// CreateTx menyimpan record payroll dalam transaksi DB.
func (r *payrollRepository) CreateTx(ctx context.Context, tx *gorm.DB, payroll *domain.Payroll) error {
	dbConn := r.db
	if tx != nil {
		dbConn = tx
	}

	// FLOW: Menjalankan kueri INSERT INTO payrolls dalam konteks transaksi
	return dbConn.WithContext(ctx).Create(payroll).Error
}
