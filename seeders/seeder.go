package seeders

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"go-jc-challange2/config"
	"go-jc-challange2/models"
	"go-jc-challange2/utils"
)

// ============================================================================
// DATABASE SEEDER - DATA AWAL UNTUK PENGUJIAN API
// ============================================================================
// Seeder mengisi database dengan data contoh yang saling berelasi (departemen,
// jabatan, karyawan, kehadiran, dan cuti) sehingga seluruh endpoint - termasuk
// proses payroll - dapat langsung diuji tanpa input manual terlebih dahulu.

// Run menjalankan seluruh proses seeding.
//
// Seeding hanya dijalankan bila tabel departments masih kosong, supaya
// menjalankan ulang aplikasi tidak membuat data contoh berlipat ganda.
func Run() {
	var departmentCount int64
	if err := config.DB.Model(&models.Department{}).Count(&departmentCount).Error; err != nil {
		log.Fatalf("[SEEDER] Gagal memeriksa data awal: %v", err)
	}

	if departmentCount > 0 {
		log.Println("[SEEDER] Data awal sudah tersedia, proses seeding dilewati")
		return
	}

	// Seluruh seeding dibungkus satu transaksi agar database tidak berakhir
	// dengan data setengah jadi bila salah satu tahap gagal.
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		departments, err := seedDepartments(tx)
		if err != nil {
			return err
		}

		positions, err := seedPositions(tx)
		if err != nil {
			return err
		}

		employees, err := seedEmployees(tx, departments, positions)
		if err != nil {
			return err
		}

		leaveDates, err := seedLeaves(tx, employees)
		if err != nil {
			return err
		}

		return seedAttendances(tx, employees, leaveDates)
	})
	if err != nil {
		log.Fatalf("[SEEDER] Gagal mengisi data awal: %v", err)
	}

	log.Println("[SEEDER] Data awal berhasil dibuat (departemen, jabatan, karyawan, kehadiran, cuti)")
	log.Printf("[SEEDER] Uji payroll dengan: POST /api/salaries/calculate {\"period\": \"%s\"}",
		time.Now().Format(utils.LayoutPeriod))
}

// ----------------------------------------------------------------------------
// 1. SEEDER DEPARTEMEN
// ----------------------------------------------------------------------------

func seedDepartments(tx *gorm.DB) ([]models.Department, error) {
	departments := []models.Department{
		{Name: "Information Technology", Code: "DEPT-IT"},
		{Name: "Human Resource", Code: "DEPT-HR"},
		{Name: "Finance", Code: "DEPT-FIN"},
		{Name: "Marketing", Code: "DEPT-MKT"},
	}

	if err := tx.Create(&departments).Error; err != nil {
		return nil, fmt.Errorf("seed departemen: %w", err)
	}

	return departments, nil
}

// ----------------------------------------------------------------------------
// 2. SEEDER JABATAN
// ----------------------------------------------------------------------------

func seedPositions(tx *gorm.DB) ([]models.Position, error) {
	// Nilai BaseSalary di bawah menjadi acuan gaji pokok saat payroll dihitung.
	positions := []models.Position{
		{Title: "Software Engineer", BaseSalary: 9000000},
		{Title: "Senior Software Engineer", BaseSalary: 14000000},
		{Title: "HR Specialist", BaseSalary: 7500000},
		{Title: "Finance Staff", BaseSalary: 7000000},
		{Title: "Marketing Executive", BaseSalary: 8000000},
	}

	if err := tx.Create(&positions).Error; err != nil {
		return nil, fmt.Errorf("seed jabatan: %w", err)
	}

	return positions, nil
}

// ----------------------------------------------------------------------------
// 3. SEEDER KARYAWAN
// ----------------------------------------------------------------------------

func seedEmployees(tx *gorm.DB, departments []models.Department, positions []models.Position) ([]models.Employee, error) {
	quota := models.DefaultLeaveQuota

	// Indeks departments dan positions merujuk pada urutan data yang dibuat di
	// dua fungsi sebelumnya, sehingga Foreign Key selalu menunjuk baris valid.
	employees := []models.Employee{
		{
			NIK: "EMP-001", FullName: "Budi Santoso", Email: "budi.santoso@ptdika.co.id",
			DepartmentID: departments[0].ID, PositionID: positions[1].ID,
			Status: models.EmployeeStatusActive, LeaveQuota: quota, LeaveBalance: quota,
		},
		{
			NIK: "EMP-002", FullName: "Siti Rahmawati", Email: "siti.rahmawati@ptdika.co.id",
			DepartmentID: departments[0].ID, PositionID: positions[0].ID,
			Status: models.EmployeeStatusActive, LeaveQuota: quota, LeaveBalance: quota,
		},
		{
			NIK: "EMP-003", FullName: "Andi Pratama", Email: "andi.pratama@ptdika.co.id",
			DepartmentID: departments[1].ID, PositionID: positions[2].ID,
			Status: models.EmployeeStatusActive, LeaveQuota: quota, LeaveBalance: quota,
		},
		{
			NIK: "EMP-004", FullName: "Dewi Lestari", Email: "dewi.lestari@ptdika.co.id",
			DepartmentID: departments[2].ID, PositionID: positions[3].ID,
			Status: models.EmployeeStatusActive, LeaveQuota: quota, LeaveBalance: quota,
		},
		{
			NIK: "EMP-005", FullName: "Rizky Ramadhan", Email: "rizky.ramadhan@ptdika.co.id",
			DepartmentID: departments[3].ID, PositionID: positions[4].ID,
			Status: models.EmployeeStatusActive, LeaveQuota: quota, LeaveBalance: quota,
		},
		{
			// Sengaja dibuat non-aktif agar filter ?status=SUSPENDED dan aturan
			// "karyawan non-aktif tidak ikut payroll" dapat langsung diuji.
			NIK: "EMP-006", FullName: "Fajar Nugroho", Email: "fajar.nugroho@ptdika.co.id",
			DepartmentID: departments[3].ID, PositionID: positions[4].ID,
			Status: models.EmployeeStatusSuspended, LeaveQuota: quota, LeaveBalance: quota,
		},
	}

	if err := tx.Create(&employees).Error; err != nil {
		return nil, fmt.Errorf("seed karyawan: %w", err)
	}

	return employees, nil
}

// ----------------------------------------------------------------------------
// 4. SEEDER PENGAJUAN CUTI
// ----------------------------------------------------------------------------

// seedLeaves membuat contoh pengajuan cuti dengan tiga status berbeda dan
// mengembalikan himpunan tanggal cuti APPROVED per karyawan. Himpunan tersebut
// dipakai seeder kehadiran agar hari cuti tidak ikut dicatat sebagai absen.
func seedLeaves(tx *gorm.DB, employees []models.Employee) (map[uint]map[string]bool, error) {
	now := time.Now()
	// Tanggal 10-12 bulan berjalan dipakai sebagai contoh cuti yang disetujui.
	leaveStart := time.Date(now.Year(), now.Month(), 10, 0, 0, 0, 0, time.UTC)
	leaveEnd := leaveStart.AddDate(0, 0, 2)

	approvedStart := leaveStart.Format(utils.LayoutDate)
	approvedEnd := leaveEnd.Format(utils.LayoutDate)

	pendingStart := time.Date(now.Year(), now.Month(), 20, 0, 0, 0, 0, time.UTC).Format(utils.LayoutDate)
	pendingEnd := time.Date(now.Year(), now.Month(), 21, 0, 0, 0, 0, time.UTC).Format(utils.LayoutDate)

	rejectedStart := time.Date(now.Year(), now.Month(), 25, 0, 0, 0, 0, time.UTC).Format(utils.LayoutDate)
	rejectedEnd := rejectedStart

	leaves := []models.Leave{
		{
			EmployeeID: employees[0].ID, StartDate: approvedStart, EndDate: approvedEnd,
			TotalDays: 3, Reason: "Cuti tahunan keperluan keluarga",
			Status: models.LeaveStatusApproved,
		},
		{
			EmployeeID: employees[1].ID, StartDate: pendingStart, EndDate: pendingEnd,
			TotalDays: 2, Reason: "Menghadiri acara pernikahan saudara",
			Status: models.LeaveStatusPending,
		},
		{
			EmployeeID: employees[2].ID, StartDate: rejectedStart, EndDate: rejectedEnd,
			TotalDays: 1, Reason: "Keperluan pribadi mendadak",
			Status: models.LeaveStatusRejected,
		},
	}

	if err := tx.Create(&leaves).Error; err != nil {
		return nil, fmt.Errorf("seed cuti: %w", err)
	}

	// Karyawan pertama memakai 3 hari cuti, sehingga saldonya berkurang.
	err := tx.Model(&models.Employee{}).
		Where("id = ?", employees[0].ID).
		Update("leave_balance", employees[0].LeaveQuota-3).Error
	if err != nil {
		return nil, fmt.Errorf("seed saldo cuti: %w", err)
	}

	// Kumpulkan tanggal cuti yang disetujui agar dilewati seeder kehadiran.
	approvedDates := map[uint]map[string]bool{
		employees[0].ID: {},
	}
	for day := leaveStart; !day.After(leaveEnd); day = day.AddDate(0, 0, 1) {
		approvedDates[employees[0].ID][day.Format(utils.LayoutDate)] = true
	}

	return approvedDates, nil
}

// ----------------------------------------------------------------------------
// 5. SEEDER KEHADIRAN
// ----------------------------------------------------------------------------

// seedAttendances membuat catatan kehadiran bulan berjalan untuk seluruh
// karyawan ACTIVE, dari tanggal 1 sampai hari ini.
//
// Pola status dibuat deterministik (bukan acak) supaya hasil perhitungan
// payroll dapat diverifikasi ulang dengan angka yang sama setiap kali.
func seedAttendances(tx *gorm.DB, employees []models.Employee, leaveDates map[uint]map[string]bool) error {
	now := time.Now()
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var attendances []models.Attendance

	for index, employee := range employees {
		// Karyawan non-aktif tidak memiliki catatan kehadiran.
		if employee.Status != models.EmployeeStatusActive {
			continue
		}

		for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
			// Lewati akhir pekan.
			if day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
				continue
			}

			dateStr := day.Format(utils.LayoutDate)

			// Lewati hari yang sudah tercatat sebagai cuti disetujui.
			if dates, ok := leaveDates[employee.ID]; ok && dates[dateStr] {
				continue
			}

			attendance := models.Attendance{
				EmployeeID: employee.ID,
				Date:       dateStr,
			}

			// Pola contoh: setiap tanggal kelipatan 9 dianggap absen, sedangkan
			// kombinasi tertentu tanggal dan karyawan dianggap terlambat.
			switch {
			case day.Day()%9 == 0:
				attendance.Status = models.AttendanceStatusAbsent
				// Hari absen tidak memiliki jam masuk maupun jam pulang.
			case (day.Day()+index)%5 == 0:
				attendance.Status = models.AttendanceStatusLate
				attendance.CheckIn = "08:25"
				attendance.CheckOut = "17:10"
			default:
				attendance.Status = models.AttendanceStatusPresent
				attendance.CheckIn = "07:45"
				attendance.CheckOut = "17:05"
			}

			attendances = append(attendances, attendance)
		}
	}

	if len(attendances) == 0 {
		return nil
	}

	// CreateInBatches memecah penyimpanan menjadi beberapa perintah INSERT,
	// menghindari batas jumlah variabel per query pada SQLite.
	if err := tx.CreateInBatches(&attendances, 100).Error; err != nil {
		return fmt.Errorf("seed kehadiran: %w", err)
	}

	return nil
}
