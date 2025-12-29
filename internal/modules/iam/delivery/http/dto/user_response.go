package dto

import "github.com/gofrs/uuid"

// UserResponse defines the publicly exposed fields of a user.
// Note that sensitive fields like PasswordHash are omitted.
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	ClinicID uuid.UUID `json:"clinic_id"`
	Email    *string   `json:"email"`
	Phone    *string   `json:"phone_number"`
	FullName string    `json:"full_name"`
	JobTitle *string   `json:"job_title"`
	Status   string    `json:"status"`
}
