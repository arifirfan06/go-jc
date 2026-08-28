package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Payroll merepresentasikan entitas Penggajian Karyawan per Bulan.
type Payroll struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EmployeeID  uint      `gorm:"not null;index" json:"employee_id"`
	Employee    User      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	BasicSalary float64   `gorm:"type:numeric(15,2);not null" json:"basic_salary"`
	Bonus       float64   `gorm:"type:numeric(15,2);not null;default:0" json:"bonus"`
	TotalSalary float64   `gorm:"type:numeric(15,2);not null" json:"total_salary"`
	Period      string    `gorm:"type:varchar(20);not null" json:"period"` // Contoh: "2026-08"
	ProcessedAt time.Time `json:"processed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PayrollRepository mendefinisikan kontrak interface untuk menyimpan record Payroll.
type PayrollRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, payroll *Payroll) error
}

// PayrollItemRequest DTO internal untuk item request transaksi payroll
type PayrollItemRequest struct {
	EmployeeID uint    `json:"employee_id"`
	Bonus      float64 `json:"bonus"`
}

// PayrollUsecase mendefinisikan kontrak logika bisnis transaksi penggajian.
type PayrollUsecase interface {
	ProcessPayroll(ctx context.Context, items []PayrollItemRequest) ([]Payroll, error)
}
