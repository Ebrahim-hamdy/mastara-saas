package dto

import "time"

// CompleteGuestRequest is used by staff to update a guest to a registered patient.
type CompleteGuestRequest struct {
	FullName    string     `json:"full_name" binding:"required,min=2"`
	Email       *string    `json:"email" binding:"omitempty,email"`
	NationalID  *string    `json:"national_id"`
	DateOfBirth *time.Time `json:"date_of_birth"`
}
