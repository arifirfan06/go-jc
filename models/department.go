package models

import "time"

// ============================================================================
// ENTITAS 1: DEPARTMENTS (DEPARTEMEN)
// ============================================================================

// Department merepresentasikan tabel "departments" di dalam database.
// Satu departemen dapat menaungi banyak karyawan (relasi one-to-many ke Employee).
type Department struct {
	// ID adalah Primary Key yang nilainya bertambah otomatis (auto increment).
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
	// Name adalah nama departemen, misalnya "Information Technology".
	Name string `gorm:"type:varchar(100);not null" json:"name"`
	// Code adalah kode unik departemen, misalnya "DEPT-IT".
	// uniqueIndex memastikan tidak ada dua departemen dengan kode yang sama.
	Code string `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName memaksa GORM memakai nama tabel "departments" secara eksplisit,
// agar penamaan tabel tidak bergantung pada aturan pluralisasi otomatis GORM.
func (Department) TableName() string {
	return "departments"
}

// ----------------------------------------------------------------------------
// DTO (Data Transfer Object) untuk request masuk
// ----------------------------------------------------------------------------

// DepartmentRequest adalah struct penampung body JSON saat membuat/memperbarui
// departemen. Struct ini sengaja dipisahkan dari model Department agar klien
// tidak bisa mengirim field sensitif seperti ID atau CreatedAt.
//
// Tag `binding` di bawah dievaluasi oleh validator bawaan Gin:
//   - required : field wajib diisi
//   - min/max   : batas panjang karakter
//   - uppercase : seluruh huruf harus kapital (mis. "DEPT-IT")
type DepartmentRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Code string `json:"code" binding:"required,min=2,max=20,uppercase"`
}
