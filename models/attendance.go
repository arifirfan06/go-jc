package models

import "time"

// ============================================================================
// ENTITAS 4: ATTENDANCES (KEHADIRAN)
// ============================================================================

// Konstanta status kehadiran harian.
const (
	AttendanceStatusPresent = "PRESENT" // Hadir tepat waktu
	AttendanceStatusLate    = "LATE"    // Hadir tetapi terlambat
	AttendanceStatusAbsent  = "ABSENT"  // Tidak hadir tanpa keterangan
)

// WorkStartTime adalah batas jam masuk kantor. Karyawan yang melakukan check-in
// setelah jam ini otomatis berstatus LATE (terlambat).
const WorkStartTime = "08:00"

// Attendance merepresentasikan tabel "attendances", yaitu catatan kehadiran
// harian seorang karyawan. Data tabel inilah yang nanti dijumlahkan pada proses
// payroll untuk menentukan besaran tunjangan dan potongan.
type Attendance struct {
	// ID adalah Primary Key auto increment.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// EmployeeID adalah Foreign Key ke tabel employees.
	//
	// EmployeeID dan Date berbagi satu index unik gabungan
	// (idx_attendance_employee_date) sehingga satu karyawan hanya boleh punya
	// satu baris kehadiran per tanggal — mencegah check-in ganda.
	EmployeeID uint `gorm:"not null;index:idx_attendance_employee_date,unique" json:"employee_id"`
	// Date adalah tanggal kehadiran dengan format YYYY-MM-DD.
	Date string `gorm:"type:varchar(10);not null;index:idx_attendance_employee_date,unique" json:"date"`
	// CheckIn adalah jam masuk dengan format HH:MM.
	CheckIn string `gorm:"type:varchar(5)" json:"check_in"`
	// CheckOut adalah jam pulang dengan format HH:MM.
	// Bernilai string kosong selama karyawan belum melakukan check-out.
	CheckOut string `gorm:"type:varchar(5)" json:"check_out"`
	// Status kehadiran: PRESENT, LATE, atau ABSENT.
	Status string `gorm:"type:varchar(10);not null;index" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Employee adalah data detail karyawan pemilik absensi ini (relasi belongs-to).
	Employee *Employee `gorm:"foreignKey:EmployeeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"employee,omitempty"`
}

// TableName memaksa GORM memakai nama tabel "attendances".
func (Attendance) TableName() string {
	return "attendances"
}

// ----------------------------------------------------------------------------
// DTO untuk request masuk
// ----------------------------------------------------------------------------

// AttendanceCheckInRequest adalah body JSON saat karyawan melakukan check-in.
//
// Tag `datetime` memakai layout waktu ala Go:
//   - datetime=2006-01-02 memaksa format tanggal YYYY-MM-DD
//   - datetime=15:04      memaksa format jam HH:MM (24 jam)
//
// Status boleh dikosongkan; bila kosong, controller menentukannya sendiri
// dengan membandingkan CheckIn terhadap WorkStartTime.
type AttendanceCheckInRequest struct {
	EmployeeID uint   `json:"employee_id" binding:"required,gt=0"`
	Date       string `json:"date" binding:"required,datetime=2006-01-02"`
	CheckIn    string `json:"check_in" binding:"required,datetime=15:04"`
	CheckOut   string `json:"check_out" binding:"omitempty,datetime=15:04"`
	Status     string `json:"status" binding:"omitempty,oneof=PRESENT LATE ABSENT"`
}

// AttendanceCheckOutRequest adalah body JSON saat karyawan mencatat jam pulang
// melalui endpoint PATCH /api/attendances/:id/checkout.
type AttendanceCheckOutRequest struct {
	CheckOut string `json:"check_out" binding:"required,datetime=15:04"`
}
