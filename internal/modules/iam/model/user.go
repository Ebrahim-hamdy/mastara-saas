// Package model contains the core domain models for the IAM module.
package model

import (
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/gofrs/uuid"
)

// User represents a staff member within a clinic.
type User struct {
	ID           uuid.UUID  `db:"id"`
	ClinicID     uuid.UUID  `db:"clinic_id"`
	Email        *string    `db:"email"`
	PhoneNumber  *string    `db:"phone_number"`
	PasswordHash *string    `db:"password_hash"`
	FullName     string     `db:"full_name"`
	JobTitle     *string    `db:"job_title"`
	Status       string     `db:"status"`
	LastLoginAt  *time.Time `db:"last_login_at"`
	InvitedByID  *uuid.UUID `db:"invited_by"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
	Roles        []Role     `db:"-"` // Loaded separately
}

// ToAuthPayload converts a user and their roles/permissions into a token payload.
func (u *User) ToAuthPayload(duration time.Duration) (*security.AuthPayload, error) {
	roleIDs := make([]uuid.UUID, len(u.Roles))
	permissionSet := make(map[string]struct{})
	for i, role := range u.Roles {
		roleIDs[i] = role.ID
		for _, p := range role.Permissions {
			permissionSet[p.PermissionKey] = struct{}{}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for p := range permissionSet {
		permissions = append(permissions, p)
	}

	return security.NewAuthPayload(u.ID, u.ClinicID, roleIDs, permissions, duration)
}
