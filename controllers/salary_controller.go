package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-jc-challange2/config"
	"go-jc-challange2/models"
	"go-jc-challange2/utils"
)

// ============================================================================
// CONTROLLER PAYROLL - PERHITUNGAN GAJI BULANAN
// ============================================================================

// attendanceRecap menampung hasil rekapitulasi kehadiran satu karyawan pada
// satu periode gaji.
type attendanceRecap struct {
	PresentDays int // Jumlah hari hadir tepat waktu
	LateDays    int // Jumlah hari terlambat
	AbsentDays  int // Jumlah hari absen
}

// CalculateSalaries menangani POST /api/salaries/calculate.
//
// Alur perhitungan untuk setiap karyawan:
//  1. Ambil gaji pokok dari jabatan (Position.BaseSalary) karyawan tersebut.
//  2. Rekap kehadiran periode berjalan dari tabel attendances untuk memperoleh
//     jumlah hari PRESENT, LATE, dan ABSENT.
//     Allowance  = PresentDays x Rp 50.000
//     Deductions = (LateDays x Rp 20.000) + (AbsentDays x Rp 100.000)
//  3. Hitung ulang saldo jatah cuti berdasarkan cuti berstatus APPROVED,
//     dan catat berapa hari cuti yang jatuh di dalam periode ini.
//  4. Simpan hasilnya ke tabel salaries dengan
//     NetSalary = BasicSalary + Allowance - Deductions.
//
// Bila employee_id dikirim, payroll hanya dihitung untuk karyawan tersebut.
// Bila dikosongkan, payroll dihitung untuk seluruh karyawan berstatus ACTIVE.
func CalculateSalaries(c *gin.Context) {
	var req models.SalaryCalculateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleBindingError(c, err)
		return
	}

	// Terjemahkan periode YYYY-MM menjadi rentang tanggal awal dan akhir bulan,
	// misalnya "2026-02" menjadi "2026-02-01" sampai "2026-02-28".
	startDate, endDate, err := utils.PeriodRange(req.Period)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, utils.SentenceCase(err.Error()))
		return
	}

	// --- Tentukan daftar karyawan yang akan diproses -------------------------
	var employees []models.Employee
	query := config.DB.Preload("Position").Preload("Department")

	if req.EmployeeID != nil {
		query = query.Where("id = ?", *req.EmployeeID)
	} else {
		// Karyawan SUSPENDED dan TERMINATED tidak ikut penggajian massal.
		query = query.Where("status = ?", models.EmployeeStatusActive)
	}

	if err := query.Order("id ASC").Find(&employees).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data karyawan")
		return
	}

	if len(employees) == 0 {
		if req.EmployeeID != nil {
			utils.Error(c, http.StatusNotFound, "Karyawan dengan ID tersebut tidak ditemukan")
			return
		}
		utils.Error(c, http.StatusNotFound,
			"Tidak ada karyawan berstatus ACTIVE yang dapat diproses payroll-nya")
		return
	}

	var results []models.Salary
	year := utils.YearOfPeriod(req.Period)

	// Seluruh perhitungan dibungkus satu transaksi: bila satu karyawan gagal
	// diproses, tidak ada satu pun slip gaji yang tersimpan setengah jadi.
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		for _, employee := range employees {
			// Jabatan wajib ada karena gaji pokok bersumber dari sana.
			if employee.Position == nil {
				return errors.New("data jabatan karyawan " + employee.FullName + " tidak ditemukan")
			}

			// --- Langkah 1: gaji pokok dari jabatan --------------------------
			basicSalary := employee.Position.BaseSalary

			// --- Langkah 2: rekap kehadiran ----------------------------------
			recap, err := recapAttendance(tx, employee.ID, startDate, endDate)
			if err != nil {
				return err
			}

			allowance := float64(recap.PresentDays) * models.AllowancePerPresentDay
			deductions := float64(recap.LateDays)*models.DeductionPerLateDay +
				float64(recap.AbsentDays)*models.DeductionPerAbsentDay

			netSalary := basicSalary + allowance - deductions
			// Gaji bersih tidak boleh negatif walaupun potongan sangat besar.
			if netSalary < 0 {
				netSalary = 0
			}

			// --- Langkah 3: cuti dan saldo jatah cuti -------------------------
			// Hari cuti yang sudah disetujui TIDAK dikenakan potongan absen,
			// karena hari tersebut memang tidak menghasilkan baris kehadiran.
			leaveDays, err := countApprovedLeaveDays(tx, employee.ID, startDate, endDate)
			if err != nil {
				return err
			}

			if err := recalculateLeaveBalance(tx, employee.ID, year); err != nil {
				return err
			}

			// Baca ulang saldo terbaru untuk disimpan sebagai bagian slip gaji.
			var refreshed models.Employee
			if err := tx.First(&refreshed, employee.ID).Error; err != nil {
				return err
			}

			// --- Langkah 4: simpan slip gaji ----------------------------------
			salary := models.Salary{
				EmployeeID:   employee.ID,
				Period:       req.Period,
				BasicSalary:  basicSalary,
				Allowance:    allowance,
				Deductions:   deductions,
				NetSalary:    netSalary,
				PresentDays:  recap.PresentDays,
				LateDays:     recap.LateDays,
				AbsentDays:   recap.AbsentDays,
				LeaveDays:    leaveDays,
				LeaveBalance: refreshed.LeaveBalance,
			}

			// Bila slip gaji periode ini sudah pernah dibuat, baris lama
			// diperbarui alih-alih dibuat ganda. Endpoint ini karenanya aman
			// dijalankan berulang kali untuk periode yang sama.
			var existing models.Salary
			findErr := tx.Where("employee_id = ? AND period = ?", employee.ID, req.Period).
				First(&existing).Error

			switch {
			case findErr == nil:
				salary.ID = existing.ID
				salary.CreatedAt = existing.CreatedAt
				if err := tx.Save(&salary).Error; err != nil {
					return err
				}
			case errors.Is(findErr, gorm.ErrRecordNotFound):
				if err := tx.Create(&salary).Error; err != nil {
					return err
				}
			default:
				return findErr
			}

			results = append(results, salary)
		}
		return nil
	})
	if err != nil {
		utils.ErrorWithDetail(c, http.StatusInternalServerError,
			"Gagal menjalankan proses payroll", err.Error())
		return
	}

	// Muat ulang seluruh slip gaji beserta relasi karyawannya agar response
	// menampilkan nama, departemen, dan jabatan, bukan sekadar employee_id.
	ids := make([]uint, 0, len(results))
	for _, salary := range results {
		ids = append(ids, salary.ID)
	}

	var salaries []models.Salary
	config.DB.
		Preload("Employee").
		Preload("Employee.Department").
		Preload("Employee.Position").
		Where("id IN ?", ids).
		Order("id ASC").
		Find(&salaries)

	// Hitung total pengeluaran gaji periode ini sebagai ringkasan bagi HR.
	var totalNet float64
	for _, salary := range salaries {
		totalNet += salary.NetSalary
	}

	utils.SuccessWithMeta(c, http.StatusOK,
		"Proses payroll periode "+req.Period+" berhasil dijalankan", salaries, gin.H{
			"period":             req.Period,
			"rentang_tanggal":    startDate + " s/d " + endDate,
			"jumlah_karyawan":    len(salaries),
			"total_gaji_dibayar": totalNet,
			"aturan_tunjangan":   models.AllowancePerPresentDay,
			"aturan_potongan":    gin.H{"terlambat": models.DeductionPerLateDay, "absen": models.DeductionPerAbsentDay},
		})
}

// GetSalariesByPeriod menangani GET /api/salaries/period/:period.
// Menampilkan seluruh slip gaji karyawan pada satu bulan tertentu.
func GetSalariesByPeriod(c *gin.Context) {
	period := c.Param("period")

	// Periode berasal dari path, bukan body JSON, sehingga tidak melewati tag
	// binding dan perlu divalidasi manual di sini.
	if _, err := utils.ParsePeriod(period); err != nil {
		utils.Error(c, http.StatusBadRequest,
			"Parameter periode harus mengikuti format YYYY-MM, contoh: 2026-08")
		return
	}

	var salaries []models.Salary
	err := config.DB.
		Preload("Employee").
		Preload("Employee.Department").
		Preload("Employee.Position").
		Where("period = ?", period).
		Order("employee_id ASC").
		Find(&salaries).Error
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Gagal mengambil data slip gaji")
		return
	}

	if len(salaries) == 0 {
		utils.SuccessWithMeta(c, http.StatusOK,
			"Belum ada slip gaji pada periode "+period+
				". Jalankan POST /api/salaries/calculate terlebih dahulu",
			[]models.Salary{}, gin.H{"period": period, "total": 0})
		return
	}

	// Ringkasan agregat untuk kebutuhan laporan HR.
	var totalNet, totalAllowance, totalDeductions float64
	for _, salary := range salaries {
		totalNet += salary.NetSalary
		totalAllowance += salary.Allowance
		totalDeductions += salary.Deductions
	}

	utils.SuccessWithMeta(c, http.StatusOK,
		"Daftar slip gaji periode "+period+" berhasil diambil", salaries, gin.H{
			"period":            period,
			"total":             len(salaries),
			"total_tunjangan":   totalAllowance,
			"total_potongan":    totalDeductions,
			"total_gaji_bersih": totalNet,
		})
}

// ============================================================================
// FUNGSI PENDUKUNG PERHITUNGAN PAYROLL
// ============================================================================

// recapAttendance menghitung jumlah hari PRESENT, LATE, dan ABSENT seorang
// karyawan di dalam rentang tanggal periode gaji.
//
// Query memakai GROUP BY status sehingga cukup sekali jalan ke database untuk
// memperoleh ketiga angka sekaligus.
func recapAttendance(tx *gorm.DB, employeeID uint, startDate, endDate string) (attendanceRecap, error) {
	// Struct anonim penampung hasil query agregat.
	var rows []struct {
		Status string
		Total  int
	}

	err := tx.Model(&models.Attendance{}).
		Select("status, COUNT(*) AS total").
		Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, startDate, endDate).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return attendanceRecap{}, err
	}

	var recap attendanceRecap
	for _, row := range rows {
		switch row.Status {
		case models.AttendanceStatusPresent:
			recap.PresentDays = row.Total
		case models.AttendanceStatusLate:
			recap.LateDays = row.Total
		case models.AttendanceStatusAbsent:
			recap.AbsentDays = row.Total
		}
	}

	return recap, nil
}

// countApprovedLeaveDays menghitung berapa hari cuti berstatus APPROVED yang
// benar-benar jatuh di dalam periode gaji.
//
// Cuti yang melintasi batas bulan (misalnya 30 Januari sampai 3 Februari)
// dipotong sesuai irisannya, sehingga hanya hari yang berada di dalam periode
// yang ikut terhitung.
func countApprovedLeaveDays(tx *gorm.DB, employeeID uint, startDate, endDate string) (int, error) {
	var leaves []models.Leave

	err := tx.Where("employee_id = ? AND status = ?", employeeID, models.LeaveStatusApproved).
		Where("start_date <= ? AND end_date >= ?", endDate, startDate).
		Find(&leaves).Error
	if err != nil {
		return 0, err
	}

	total := 0
	for _, leave := range leaves {
		total += utils.CountOverlapDays(leave.StartDate, leave.EndDate, startDate, endDate)
	}

	return total, nil
}
