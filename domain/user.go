package domain

import (
	"context"
	"time"
)

// Const Role Karyawan & Admin HRD
const (
	RoleHRD      = "HRD"
	RoleEmployee = "EMPLOYEE"
)

// User merepresentasikan entitas Pengguna / Karyawan dalam sistem HRIS.
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"type:varchar(100);not null" json:"name"`
	Email        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password     string     `gorm:"type:varchar(255);not null" json:"-"`
	Role         string     `gorm:"type:varchar(20);not null;default:'EMPLOYEE'" json:"role"`
	DepartmentID uint       `gorm:"not null" json:"department_id"`
	Department   Department `gorm:"foreignKey:DepartmentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"department,omitempty"`
	BasicSalary  float64    `gorm:"type:numeric(15,2);not null;default:0" json:"basic_salary"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserRepository mendefinisikan kontrak interface untuk interaksi data User di database.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

// AuthUsecase mendefinisikan kontrak logika bisnis untuk Autentikasi.
type AuthUsecase interface {
	Register(ctx context.Context, name, email, password, role string, departmentID uint, basicSalary float64) (*User, string, error)
	Login(ctx context.Context, email, password string) (*User, string, error)
	Logout(ctx context.Context, tokenString string) error
}

