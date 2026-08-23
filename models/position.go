package models

import "time"

// ============================================================================
// ENTITAS 2: POSITIONS (JABATAN)
// ============================================================================

// Position merepresentasikan tabel "positions".
// BaseSalary pada jabatan menjadi acuan gaji pokok saat proses payroll berjalan.
type Position struct {
	// ID adalah Primary Key auto increment.
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// Title adalah nama jabatan, misalnya "Software Engineer".
	Title string `gorm:"type:varchar(100);not null;uniqueIndex" json:"title"`
	// BaseSalary adalah gaji pokok bawaan untuk jabatan tersebut (dalam Rupiah).
	BaseSalary float64 `gorm:"not null;default:0" json:"base_salary"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName memaksa GORM memakai nama tabel "positions".
func (Position) TableName() string {
	return "positions"
}

// ----------------------------------------------------------------------------
// DTO untuk request masuk
// ----------------------------------------------------------------------------

// PositionRequest adalah penampung body JSON untuk membuat/memperbarui jabatan.
//
// BaseSalary sengaja hanya memakai `gt=0` tanpa `required`. Bagi validator, nilai
// nol pada tipe angka dianggap sama dengan "kosong", sehingga `required` akan
// menghasilkan pesan "wajib diisi" ketika klien benar-benar mengirim angka 0 -
// sebuah pesan yang menyesatkan. Dengan `gt=0`, field yang tidak dikirim maupun
// yang bernilai 0 sama-sama menerima pesan "harus lebih besar dari 0".
type PositionRequest struct {
	Title      string  `json:"title" binding:"required,min=3,max=100"`
	BaseSalary float64 `json:"base_salary" binding:"gt=0"`
}
