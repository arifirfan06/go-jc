package repository

import (
	"context"
	"errors"
	"go-hris-payroll-system/domain"

	"gorm.io/gorm"
)

type leaveRepository struct {
	db *gorm.DB
}

// NewLeaveRepository menginisialisasi repository leave.
func NewLeaveRepository(db *gorm.DB) domain.LeaveRepository {
	return &leaveRepository{db: db}
}

// Create menyimpan pengajuan cuti baru.
func (r *leaveRepository) Create(ctx context.Context, leave *domain.Leave) error {
	// FLOW: Menjalankan kueri INSERT INTO leaves
	return r.db.WithContext(ctx).Create(leave).Error
}

// FindByIDWithDefensiveCheck mengambil detail pengajuan cuti dengan Mitigasi IDOR defensif.
func (r *leaveRepository) FindByIDWithDefensiveCheck(ctx context.Context, id uint, filterEmployeeID uint) (*domain.Leave, error) {
	var leave domain.Leave
	query := r.db.WithContext(ctx).Preload("Employee")

	// FLOW 1: Mitigasi IDOR Defensif.
	// Jika filterEmployeeID > 0 (Akses oleh EMPLOYEE), tambahkan syarat Kueri: WHERE id = ? AND employee_id = ?
	// Karyawan tidak bisa melihat data cuti karyawan lain meskipun berhasil menebak ID cutinya.
	if filterEmployeeID > 0 {
		query = query.Where("id = ? AND employee_id = ?", id, filterEmployeeID)
	} else {
		// Jika filterEmployeeID == 0 (Akses oleh HRD), ambil data tanpa pembatasan employee_id
		query = query.Where("id = ?", id)
	}

	// FLOW 2: Eksekusi kueri ke database
	err := query.First(&leave).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &leave, nil
}

// FindAllByEmployeeID mengambil daftar seluruh cuti milik 1 karyawan tertentu.
func (r *leaveRepository) FindAllByEmployeeID(ctx context.Context, employeeID uint) ([]domain.Leave, error) {
	var leaves []domain.Leave
	// FLOW: Menjalankan kueri SELECT * FROM leaves WHERE employee_id = ?
	err := r.db.WithContext(ctx).Preload("Employee").Where("employee_id = ?", employeeID).Order("created_at desc").Find(&leaves).Error
	return leaves, err
}

// FindAll mengambil daftar seluruh pengajuan cuti seluruh karyawan (Untuk HRD).
func (r *leaveRepository) FindAll(ctx context.Context) ([]domain.Leave, error) {
	var leaves []domain.Leave
	// FLOW: Menjalankan kueri SELECT * FROM leaves (HRD view)
	err := r.db.WithContext(ctx).Preload("Employee").Order("created_at desc").Find(&leaves).Error
	return leaves, err
}
