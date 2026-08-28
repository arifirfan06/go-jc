package usecase

import (
	"context"
	"fmt"
	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"
	"time"
)

type leaveUsecase struct {
	leaveRepo domain.LeaveRepository
	userRepo  domain.UserRepository
}

// NewLeaveUsecase menginisialisasi usecase pengajuan cuti.
func NewLeaveUsecase(leaveRepo domain.LeaveRepository, userRepo domain.UserRepository) domain.LeaveUsecase {
	return &leaveUsecase{
		leaveRepo: leaveRepo,
		userRepo:  userRepo,
	}
}

// CreateLeave memproses pengajuan cuti baru oleh karyawan.
func (u *leaveUsecase) CreateLeave(ctx context.Context, employeeID uint, reason string, startDate, endDate time.Time) (*domain.Leave, error) {
	// FLOW 1: Validasi tanggal mulai tidak boleh lebih besar dari tanggal selesai
	if startDate.After(endDate) {
		return nil, fmt.Errorf("tanggal mulai cuti tidak boleh melebihi tanggal selesai")
	}

	// FLOW 2: Buat entitas Leave baru dengan status awal PENDING
	leave := &domain.Leave{
		EmployeeID: employeeID,
		Reason:     reason,
		StartDate:  startDate,
		EndDate:    endDate,
		Status:     domain.LeaveStatusPending,
	}

	// FLOW 3: Simpan pengajuan cuti ke database
	if err := u.leaveRepo.Create(ctx, leave); err != nil {
		return nil, fmt.Errorf("gagal mengajukan cuti: %w", err)
	}

	return leave, nil
}

// GetLeaveByID mengambil detail cuti berdasarkan ID dengan Mitigasi IDOR berdasarkan Role Pengakses.
func (u *leaveUsecase) GetLeaveByID(ctx context.Context, id uint, requesterID uint, requesterRole string) (*domain.Leave, error) {
	var filterEmployeeID uint

	// FLOW 1: Terapkan Penanganan IDOR Defensif berdasarkan Role:
	// - Jika requesterRole == "EMPLOYEE", saring hanya cuti milik requesterID tersebut
	// - Jika requesterRole == "HRD", set filter 0 (dapat melihat cuti seluruh karyawan)
	if requesterRole == domain.RoleEmployee {
		filterEmployeeID = requesterID
	} else {
		filterEmployeeID = 0
	}

	// FLOW 2: Panggil Repository dengan kueri defensif
	leave, err := u.leaveRepo.FindByIDWithDefensiveCheck(ctx, id, filterEmployeeID)
	if err != nil {
		return nil, err
	}

	return leave, nil
}

// GetLeaves mengambil daftar riwayat cuti (Seluruh cuti untuk HRD, atau cuti pribadi untuk EMPLOYEE).
func (u *leaveUsecase) GetLeaves(ctx context.Context, requesterID uint, requesterRole string) ([]domain.Leave, error) {
	// FLOW: Mengambil daftar cuti sesuai hak akses role
	if requesterRole == domain.RoleHRD {
		return u.leaveRepo.FindAll(ctx)
	}
	return u.leaveRepo.FindAllByEmployeeID(ctx, requesterID)
}

// ConvertToLeaveResponse helper konversi entitas Leave ke DTO LeaveResponse
func ConvertToLeaveResponse(l *domain.Leave) dto.LeaveResponse {
	return dto.LeaveResponse{
		ID:         l.ID,
		EmployeeID: l.EmployeeID,
		Reason:     l.Reason,
		StartDate:  l.StartDate.Format("2006-01-02"),
		EndDate:    l.EndDate.Format("2006-01-02"),
		Status:     l.Status,
		CreatedAt:  l.CreatedAt,
	}
}

// ConvertToLeaveResponses helper konversi slice entitas Leave ke DTO LeaveResponse
func ConvertToLeaveResponses(leaves []domain.Leave) []dto.LeaveResponse {
	var responses []dto.LeaveResponse
	for _, l := range leaves {
		responses = append(responses, ConvertToLeaveResponse(&l))
	}
	return responses
}
