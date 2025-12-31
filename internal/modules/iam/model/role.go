package model

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a collection of permissions.
type Role struct {
	ID           uuid.UUID    `db:"id"`
	ClinicID     *uuid.UUID   `db:"clinic_id"` // Null for system roles
	Name         string       `db:"name"`
	Description  *string      `db:"description"`
	IsSystemRole bool         `db:"is_system_role"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
	Permissions  []Permission `db:"-"` // Loaded separately
}
