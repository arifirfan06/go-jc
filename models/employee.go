package models

import "time"

// ============================================================================
// ENTITAS 3: EMPLOYEES (KARYAWAN)
// ============================================================================

// Konstanta status karyawan. Dipakai bersama oleh controller dan seeder agar
// tidak ada salah ketik string ("ACTIVE" vs "Active") yang tersebar di kode.
const (
	EmployeeStatusActive     = "ACTIVE"     // Karyawan aktif bekerja
	EmployeeStatusSuspended  = "SUSPENDED"  // Karyawan diskors sementara
	EmployeeStatusTerminated = "TERMINATED" // Karyawan sudah tidak bekerja
)

// DefaultLeaveQuota adalah jatah cuti tahunan bawaan (12 hari kerja per tahun)
// yang diberikan kepada karyawan baru bila tidak ditentukan secara eksplisit.
const DefaultLeaveQuota = 12

// Employee merepresentasikan tabel "employees".
// Tabel ini menjadi pusat relasi: menunjuk ke Departments dan Positions
// (belongs-to), serta dirujuk oleh Attendances, Leaves, dan Salaries.
type Employee struct {
	// ID adalah Primary Key auto increment.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// NIK adalah Nomor Induk Karyawan yang bersifat unik.
	NIK string `gorm:"column:nik;type:varchar(30);not null;uniqueIndex" json:"nik"`
	// FullName adalah nama lengkap karyawan.
	FullName string `gorm:"type:varchar(100);not null;index" json:"full_name"`
	// Email adalah alamat surel karyawan, unik dan wajib berformat email valid.
	Email string `gorm:"type:varchar(120);not null;uniqueIndex" json:"email"`

	// DepartmentID adalah Foreign Key yang menunjuk ke tabel departments.
	// constraint OnDelete:RESTRICT mencegah departemen dihapus selama masih
	// dipakai oleh karyawan, sehingga integritas relasi tetap terjaga.
	DepartmentID uint `gorm:"not null;index" json:"department_id"`
	// PositionID adalah Foreign Key yang menunjuk ke tabel positions.
	PositionID uint `gorm:"not null;index" json:"position_id"`

	// Status karyawan: ACTIVE, SUSPENDED, atau TERMINATED.
	Status string `gorm:"type:varchar(15);not null;default:ACTIVE;index" json:"status"`

	// LeaveQuota adalah jatah cuti tahunan (hari) milik karyawan.
	LeaveQuota int `gorm:"not null;default:12" json:"leave_quota"`
	// LeaveBalance adalah sisa jatah cuti berjalan. Nilainya dihitung ulang
	// (LeaveQuota dikurangi total hari cuti APPROVED tahun berjalan) setiap kali
	// cuti disetujui maupun saat proses payroll dijalankan.
	LeaveBalance int `gorm:"not null;default:12" json:"leave_balance"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// ------------------------------------------------------------------
	// Relasi (diisi hanya ketika controller memanggil Preload)
	// ------------------------------------------------------------------

	// Department adalah data detail departemen karyawan (relasi belongs-to).
	Department *Department `gorm:"foreignKey:DepartmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"department,omitempty"`
	// Position adalah data detail jabatan karyawan (relasi belongs-to).
	Position *Position `gorm:"foreignKey:PositionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"position,omitempty"`
}

// TableName memaksa GORM memakai nama tabel "employees".
func (Employee) TableName() string {
	return "employees"
}

// ----------------------------------------------------------------------------
// DTO untuk request masuk
// ----------------------------------------------------------------------------

// EmployeeRequest adalah penampung body JSON saat menambah karyawan baru.
//
// Catatan validasi:
//   - `email`  memastikan format alamat surel valid (mis. budi@perusahaan.co.id)
//   - `oneof`  membatasi Status hanya pada tiga nilai yang diizinkan
//   - Status memakai `omitempty`, artinya boleh dikosongkan dan otomatis
//     diisi ACTIVE oleh controller.
type EmployeeRequest struct {
	NIK          string `json:"nik" binding:"required,min=3,max=30"`
	FullName     string `json:"full_name" binding:"required,min=3,max=100"`
	Email        string `json:"email" binding:"required,email,max=120"`
	DepartmentID uint   `json:"department_id" binding:"required,gt=0"`
	PositionID   uint   `json:"position_id" binding:"required,gt=0"`
	Status       string `json:"status" binding:"omitempty,oneof=ACTIVE SUSPENDED TERMINATED"`
	// LeaveQuota memakai pointer agar bisa dibedakan antara "tidak dikirim" (nil)
	// dan "dikirim dengan nilai 0". Bila nil, dipakai DefaultLeaveQuota.
	LeaveQuota *int `json:"leave_quota" binding:"omitempty,gte=0,lte=60"`
}

// EmployeeUpdateRequest adalah penampung body JSON saat memperbarui karyawan.
// Seluruh field memakai pointer + `omitempty` sehingga klien cukup mengirim
// field yang ingin diubah saja (partial update), termasuk memindahkan karyawan
// ke departemen atau jabatan lain.
type EmployeeUpdateRequest struct {
	NIK          *string `json:"nik" binding:"omitempty,min=3,max=30"`
	FullName     *string `json:"full_name" binding:"omitempty,min=3,max=100"`
	Email        *string `json:"email" binding:"omitempty,email,max=120"`
	DepartmentID *uint   `json:"department_id" binding:"omitempty,gt=0"`
	PositionID   *uint   `json:"position_id" binding:"omitempty,gt=0"`
	Status       *string `json:"status" binding:"omitempty,oneof=ACTIVE SUSPENDED TERMINATED"`
	LeaveQuota   *int    `json:"leave_quota" binding:"omitempty,gte=0,lte=60"`
}
