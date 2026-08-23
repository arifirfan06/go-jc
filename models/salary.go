package models

import "time"

// ============================================================================
// ENTITAS 6: SALARIES (SLIP GAJI / PAYROLL)
// ============================================================================

// Aturan perhitungan payroll (nilai dalam Rupiah).
const (
	// AllowancePerPresentDay adalah bonus tunjangan kehadiran untuk setiap hari
	// karyawan hadir tepat waktu (status PRESENT).
	AllowancePerPresentDay = 50000.0
	// DeductionPerLateDay adalah potongan untuk setiap hari terlambat (LATE).
	DeductionPerLateDay = 20000.0
	// DeductionPerAbsentDay adalah potongan untuk setiap hari absen (ABSENT).
	DeductionPerAbsentDay = 100000.0
)

// Salary merepresentasikan tabel "salaries", yaitu slip gaji hasil rekapitulasi
// payroll seorang karyawan pada satu periode (satu bulan).
type Salary struct {
	// ID adalah Primary Key auto increment.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// EmployeeID adalah Foreign Key ke tabel employees.
	//
	// EmployeeID dan Period berbagi index unik gabungan
	// (idx_salary_employee_period) supaya satu karyawan hanya punya satu slip
	// gaji per periode. Menjalankan ulang payroll akan memperbarui baris yang
	// sudah ada, bukan membuat duplikat.
	EmployeeID uint `gorm:"not null;index:idx_salary_employee_period,unique" json:"employee_id"`
	// Period adalah bulan dan tahun periode gaji dengan format YYYY-MM.
	Period string `gorm:"type:varchar(7);not null;index:idx_salary_employee_period,unique" json:"period"`

	// BasicSalary adalah gaji pokok saat pembayaran, disalin dari BaseSalary
	// jabatan karyawan. Disimpan sebagai snapshot agar slip gaji lama tetap
	// menampilkan angka yang benar walaupun gaji jabatan diubah kemudian.
	BasicSalary float64 `gorm:"not null;default:0" json:"basic_salary"`
	// Allowance adalah total tunjangan kehadiran periode ini.
	Allowance float64 `gorm:"not null;default:0" json:"allowance"`
	// Deductions adalah total potongan (keterlambatan + absen) periode ini.
	Deductions float64 `gorm:"not null;default:0" json:"deductions"`
	// NetSalary adalah gaji bersih akhir: BasicSalary + Allowance - Deductions.
	NetSalary float64 `gorm:"not null;default:0" json:"net_salary"`

	// ------------------------------------------------------------------
	// Kolom ringkasan pendukung: menjelaskan asal-usul angka di atas
	// sehingga slip gaji dapat diaudit tanpa query tambahan.
	// ------------------------------------------------------------------

	// PresentDays adalah jumlah hari hadir tepat waktu pada periode ini.
	PresentDays int `gorm:"not null;default:0" json:"present_days"`
	// LateDays adalah jumlah hari terlambat pada periode ini.
	LateDays int `gorm:"not null;default:0" json:"late_days"`
	// AbsentDays adalah jumlah hari absen pada periode ini.
	AbsentDays int `gorm:"not null;default:0" json:"absent_days"`
	// LeaveDays adalah jumlah hari cuti APPROVED yang jatuh pada periode ini.
	// Hari cuti yang disetujui tidak dikenakan potongan absen.
	LeaveDays int `gorm:"not null;default:0" json:"leave_days"`
	// LeaveBalance adalah sisa jatah cuti karyawan setelah dikurangi seluruh
	// cuti APPROVED pada tahun periode ini.
	LeaveBalance int `gorm:"not null;default:0" json:"leave_balance"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Employee adalah data detail karyawan pemilik slip gaji (relasi belongs-to).
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"employee,omitempty"`
}

// TableName memaksa GORM memakai nama tabel "salaries".
func (Salary) TableName() string {
	return "salaries"
}

// ----------------------------------------------------------------------------
// DTO untuk request masuk
// ----------------------------------------------------------------------------

// SalaryCalculateRequest adalah body JSON untuk memicu proses payroll bulanan.
//
//   - Period wajib berformat YYYY-MM (divalidasi lewat layout Go "2006-01").
//   - EmployeeID bersifat opsional: bila diisi, payroll hanya dihitung untuk
//     satu karyawan; bila dikosongkan, payroll dihitung untuk seluruh karyawan
//     berstatus ACTIVE.
type SalaryCalculateRequest struct {
	Period     string `json:"period" binding:"required,datetime=2006-01"`
	EmployeeID *uint  `json:"employee_id" binding:"omitempty,gt=0"`
}
