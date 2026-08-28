package dto

import "time"

// CreateLeaveRequest DTO payload pengajuan cuti karyawan.
type CreateLeaveRequest struct {
	Reason    string `json:"reason" binding:"required"`
	StartDate string `json:"start_date" binding:"required"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // Format: YYYY-MM-DD
}

// LeaveResponse DTO balasan detail data cuti.
type LeaveResponse struct {
	ID         uint      `json:"id"`
	EmployeeID uint      `json:"employee_id"`
	Reason     string    `json:"reason"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
