package dto

// RegisterRequest DTO payload untuk registrasi karyawan/HRD baru.
type RegisterRequest struct {
	Name         string  `json:"name" binding:"required"`
	Email        string  `json:"email" binding:"required,email,company_email"` // Menggunakan custom validator company_email (@company.co.id)
	Password     string  `json:"password" binding:"required,min=6"`
	Role         string  `json:"role" binding:"required,oneof=HRD EMPLOYEE"`
	DepartmentID uint    `json:"department_id" binding:"required"`
	BasicSalary  float64 `json:"basic_salary" binding:"required,gte=0"`
}

// LoginRequest DTO payload untuk login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse DTO balasan data autentikasi beserta JWT token.
type AuthResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Token string `json:"token"`
}
