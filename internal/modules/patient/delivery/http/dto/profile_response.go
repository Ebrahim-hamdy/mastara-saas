package dto

import (
	"time"

	"github.com/gofrs/uuid"
)

// ProfileResponse defines the publicly exposed fields of a patient profile.
type ProfileResponse struct {
	ID            uuid.UUID  `json:"id"`
	ClinicID      uuid.UUID  `json:"clinic_id"`
	FullName      string     `json:"full_name"`
	PhoneNumber   *string    `json:"phone_number"`
	Email         *string    `json:"email"`
	NationalID    *string    `json:"national_id"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	ProfileStatus string     `json:"profile_status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
