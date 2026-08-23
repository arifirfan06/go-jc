package models

import "time"

// ============================================================================
// ENTITAS 5: LEAVES (PENGAJUAN CUTI)
// ============================================================================

// Konstanta status persetujuan cuti.
const (
	LeaveStatusPending  = "PENDING"  // Menunggu keputusan HR
	LeaveStatusApproved = "APPROVED" // Disetujui HR
	LeaveStatusRejected = "REJECTED" // Ditolak HR
)

// Leave merepresentasikan tabel "leaves", yaitu pengajuan cuti karyawan.
// Setiap pengajuan baru selalu dimulai dari status PENDING, lalu diubah oleh HR
// menjadi APPROVED atau REJECTED melalui endpoint approval.
type Leave struct {
	// ID adalah Primary Key auto increment.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// EmployeeID adalah Foreign Key ke tabel employees.
	EmployeeID uint `gorm:"not null;index" json:"employee_id"`
	// StartDate adalah tanggal mulai cuti (format YYYY-MM-DD).
	StartDate string `gorm:"type:varchar(10);not null;index" json:"start_date"`
	// EndDate adalah tanggal selesai cuti (format YYYY-MM-DD), inklusif.
	EndDate string `gorm:"type:varchar(10);not null;index" json:"end_date"`
	// TotalDays adalah jumlah hari cuti (EndDate - StartDate + 1). Disimpan agar
	// pengurangan saldo cuti tidak perlu menghitung ulang selisih tanggal.
	TotalDays int `gorm:"not null;default:0" json:"total_days"`
	// Reason adalah alasan pengajuan cuti.
	Reason string `gorm:"type:varchar(255);not null" json:"reason"`
	// Status persetujuan: PENDING, APPROVED, atau REJECTED.
	Status string `gorm:"type:varchar(10);not null;default:PENDING;index" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Employee adalah data detail karyawan pengaju cuti (relasi belongs-to).
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"employee,omitempty"`
}

// TableName memaksa GORM memakai nama tabel "leaves".
func (Leave) TableName() string {
	return "leaves"
}

// ----------------------------------------------------------------------------
// DTO untuk request masuk
// ----------------------------------------------------------------------------

// LeaveRequest adalah body JSON saat karyawan mengajukan cuti baru.
// Urutan tanggal (EndDate tidak boleh mendahului StartDate) divalidasi di
// controller karena membutuhkan perbandingan antar-field.
type LeaveRequest struct {
	EmployeeID uint   `json:"employee_id" binding:"required,gt=0"`
	StartDate  string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate    string `json:"end_date" binding:"required,datetime=2006-01-02"`
	Reason     string `json:"reason" binding:"required,min=5,max=255"`
}

// LeaveApprovalRequest adalah body JSON yang dikirim HR untuk menyetujui atau
// menolak pengajuan cuti. `oneof` menutup kemungkinan status dikembalikan ke
// PENDING lewat endpoint approval.
type LeaveApprovalRequest struct {
	Status string `json:"status" binding:"required,oneof=APPROVED REJECTED"`
}
