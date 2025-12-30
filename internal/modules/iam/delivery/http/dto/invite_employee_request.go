package dto

// InviteEmployeeRequest defines the API contract for inviting a new employee.
type InviteEmployeeRequest struct {
	FullName    string  `json:"full_name"`
	Email       *string `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	JobTitle    *string `json:"job_title"`
}
