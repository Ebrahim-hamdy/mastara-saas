// Package model contains the core domain models for the Patient module, centered around the Profile entity.
package model

import (
	"time"

	"github.com/google/uuid"
)

// ProfileStatus represents the state of a patient's profile.
type ProfileStatus string

const (
	ProfileStatusGuest      ProfileStatus = "GUEST"
	ProfileStatusRegistered ProfileStatus = "REGISTERED"
	ProfileStatusArchived   ProfileStatus = "ARCHIVED"
)

// Profile represents an individual in the system, who can be a patient.
// This struct maps directly to the 'profiles' table.
type Profile struct {
	ID            uuid.UUID     `db:"id"`
	ClinicID      uuid.UUID     `db:"clinic_id"`
	FullName      string        `db:"full_name"`
	PhoneNumber   *string       `db:"phone_number"`
	Email         *string       `db:"email"`
	NationalID    *string       `db:"national_id"`
	DateOfBirth   *time.Time    `db:"date_of_birth"`
	ProfileStatus ProfileStatus `db:"profile_status"`
	ExtendedData  []byte        `db:"extended_data"` // Stays as []byte for raw JSONB
	CreatedAt     time.Time     `db:"created_at"`
	UpdatedAt     time.Time     `db:"updated_at"`
	DeletedAt     *time.Time    `db:"deleted_at"`
}
