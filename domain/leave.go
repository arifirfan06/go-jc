package domain

import (
	"context"
	"time"
)

// Status Cuti Karyawan
const (
	LeaveStatusPending  = "PENDING"
	LeaveStatusApproved = "APPROVED"
	LeaveStatusRejected = "REJECTED"
)

// Leave merepresentasikan entitas Pengajuan Cuti Karyawan.
type Leave struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EmployeeID uint      `gorm:"not null;index" json:"employee_id"`
	Employee   User      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	Reason     string    `gorm:"type:text;not null" json:"reason"`
	StartDate  time.Time `gorm:"type:date;not null" json:"start_date"`
	EndDate    time.Time `gorm:"type:date;not null" json:"end_date"`
	Status     string    `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LeaveRepository mendefinisikan kontrak akses database untuk Pengajuan Cuti dengan mitigasi IDOR.
type LeaveRepository interface {
	Create(ctx context.Context, leave *Leave) error
	// FindByIDWithDefensiveCheck mengambil detail cuti.
	// Jika filterEmployeeID > 0, kueri menerapkan WHERE id = :id AND employee_id = :filterEmployeeID (Mencegah IDOR untuk EMPLOYEE).
	// Jika filterEmployeeID == 0, kueri mengambil detail tanpa membatasi employee_id (Untuk HRD).
	FindByIDWithDefensiveCheck(ctx context.Context, id uint, filterEmployeeID uint) (*Leave, error)
	FindAllByEmployeeID(ctx context.Context, employeeID uint) ([]Leave, error)
	FindAll(ctx context.Context) ([]Leave, error)
}

// LeaveUsecase mendefinisikan kontrak logika bisnis Cuti.
type LeaveUsecase interface {
	CreateLeave(ctx context.Context, employeeID uint, reason string, startDate, endDate time.Time) (*Leave, error)
	GetLeaveByID(ctx context.Context, id uint, requesterID uint, requesterRole string) (*Leave, error)
	GetLeaves(ctx context.Context, requesterID uint, requesterRole string) ([]Leave, error)
}
