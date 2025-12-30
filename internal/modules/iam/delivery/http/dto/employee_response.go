package dto

import "github.com/gofrs/uuid"

// EmployeeResponse defines the publicly exposed fields of an employee.
// It combines data from both the 'profiles' and 'employees' tables.
type EmployeeResponse struct {
	ID          uuid.UUID `json:"id"` // This is the Profile ID
	ClinicID    uuid.UUID `json:"clinic_id"`
	Email       *string   `json:"email"`
	PhoneNumber *string   `json:"phone_number"`
	FullName    string    `json:"full_name"`
	JobTitle    *string   `json:"job_title"`
	Status      string    `json:"status"`
}
