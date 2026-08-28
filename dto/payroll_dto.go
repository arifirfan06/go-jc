package dto

// PayrollProcessItem DTO item karyawan dan bonus yang diproses.
type PayrollProcessItem struct {
	EmployeeID uint    `json:"employee_id" binding:"required"`
	Bonus      float64 `json:"bonus" binding:"gte=0"`
}

// ProcessPayrollRequest DTO array list karyawan untuk diproses penggajiannya.
type ProcessPayrollRequest struct {
	Payrolls []PayrollProcessItem `json:"payrolls" binding:"required,dive"`
}

// PayrollResponse DTO detail hasil penggajian karyawan.
type PayrollResponse struct {
	ID          uint    `json:"id"`
	EmployeeID  uint    `json:"employee_id"`
	BasicSalary float64 `json:"basic_salary"`
	Bonus       float64 `json:"bonus"`
	TotalSalary float64 `json:"total_salary"`
	Period      string  `json:"period"`
	ProcessedAt string  `json:"processed_at"`
}
