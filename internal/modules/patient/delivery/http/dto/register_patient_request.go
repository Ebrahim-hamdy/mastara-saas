package dto

import "time"

// RegisterPatientRequest is used by staff for in-clinic full registration.
type RegisterPatientRequest struct {
	FullName    string     `json:"full_name" binding:"required,min=2"`
	PhoneNumber string     `json:"phone_number" binding:"required,e164"`
	Email       *string    `json:"email" binding:"omitempty,email"`
	NationalID  *string    `json:"national_id"`
	DateOfBirth *time.Time `json:"date_of_birth"`
}
