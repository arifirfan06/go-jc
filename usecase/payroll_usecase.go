package usecase

import (
	"context"
	"fmt"
	"go-hris-payroll-system/domain"
	"go-hris-payroll-system/dto"
	"time"

	"gorm.io/gorm"
)

type payrollUsecase struct {
	db             *gorm.DB
	userRepo       domain.UserRepository
	deptRepo       domain.DepartmentRepository
	payrollRepo    domain.PayrollRepository
}

// NewPayrollUsecase menginisialisasi usecase penggajian.
func NewPayrollUsecase(
	db *gorm.DB,
	userRepo domain.UserRepository,
	deptRepo domain.DepartmentRepository,
	payrollRepo domain.PayrollRepository,
) domain.PayrollUsecase {
	return &payrollUsecase{
		db:          db,
		userRepo:    userRepo,
		deptRepo:    deptRepo,
		payrollRepo: payrollRepo,
	}
}

// ProcessPayroll memproses transaksi penggajian karyawan secara atomik dengan DB Transaction & Row Locking.
func (u *payrollUsecase) ProcessPayroll(ctx context.Context, items []domain.PayrollItemRequest) ([]domain.Payroll, error) {
	var createdPayrolls []domain.Payroll
	currentPeriod := time.Now().Format("2006-01")

	// FLOW 1: Membuka Transaksi Database (db.Transaction)
	// Seluruh perhitungan, locking, dan pemotongan anggaran dijalankan di dalam blok transaksi ini.
	// Jika terjadi error di salah satu langkah, GORM akan otomatis melakukan ROLLBACK secara keseluruhan.
	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// FLOW 2: Iterasi seluruh list karyawan yang diproses penggajiannya
		for _, item := range items {
			// FLOW 2.1: Ambil data karyawan (gaji pokok & ID departemen)
			employee, err := u.userRepo.FindByID(ctx, item.EmployeeID)
			if err != nil {
				return fmt.Errorf("karyawan dengan ID %d tidak ditemukan: %w", item.EmployeeID, err)
			}

			// FLOW 2.2: Hitung Gaji Total = Gaji Pokok + Bonus
			totalSalary := employee.BasicSalary + item.Bonus

			// FLOW 2.3: Gunakan Row-Level Locking (FOR UPDATE) pada baris budget departemen
			// Mencegah race condition dari transaksi concurrent lain yang mencoba mengubah budget bersamaan
			dept, err := u.deptRepo.FindByIDWithLock(ctx, tx, employee.DepartmentID)
			if err != nil {
				return fmt.Errorf("gagal mengambil budget departemen (ID: %d): %w", employee.DepartmentID, err)
			}

			// FLOW 2.4: Validasi kecukupan anggaran departemen
			if dept.Budget < totalSalary {
				// Membatalkan (Rollback) transaksi dan mengembalikan error informatif
				return fmt.Errorf(
					"%w: Anggaran departemen '%s' tidak mencukupi. Sisa budget: Rp%.2f, Dibutuhkan untuk Karyawan %s (ID: %d): Rp%.2f",
					domain.ErrInsufficientBudget, dept.Name, dept.Budget, employee.Name, employee.ID, totalSalary,
				)
			}

			// FLOW 2.5: Potong Anggaran Departemen dalam transaksi
			if err := u.deptRepo.DeductBudget(ctx, tx, employee.DepartmentID, totalSalary); err != nil {
				return fmt.Errorf("gagal memotong anggaran departemen '%s': %w", dept.Name, err)
			}

			// FLOW 2.6: Buat record Payroll baru
			payrollRecord := &domain.Payroll{
				EmployeeID:  employee.ID,
				BasicSalary: employee.BasicSalary,
				Bonus:       item.Bonus,
				TotalSalary: totalSalary,
				Period:      currentPeriod,
				ProcessedAt: time.Now(),
			}

			if err := u.payrollRepo.CreateTx(ctx, tx, payrollRecord); err != nil {
				return fmt.Errorf("gagal menyimpan record penggajian untuk karyawan ID %d: %w", employee.ID, err)
			}

			payrollRecord.Employee = *employee
			createdPayrolls = append(createdPayrolls, *payrollRecord)
		}

		// FLOW 3: Jika seluruh loop berhasil tanpa error, Transaksi akan COMMIT otomatis
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdPayrolls, nil
}

// ConvertToPayrollResponses helper untuk mengonversi entitas ke DTO response
func ConvertToPayrollResponses(payrolls []domain.Payroll) []dto.PayrollResponse {
	var responses []dto.PayrollResponse
	for _, p := range payrolls {
		responses = append(responses, dto.PayrollResponse{
			ID:          p.ID,
			EmployeeID:  p.EmployeeID,
			BasicSalary: p.BasicSalary,
			Bonus:       p.Bonus,
			TotalSalary: p.TotalSalary,
			Period:      p.Period,
			ProcessedAt: p.ProcessedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return responses
}
