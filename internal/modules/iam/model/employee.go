package model

import (
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/google/uuid"
)

type EmployeeStatus string

const (
	EmployeeStatusInvited    EmployeeStatus = "INVITED"
	EmployeeStatusActive     EmployeeStatus = "ACTIVE"
	EmployeeStatusSuspended  EmployeeStatus = "SUSPENDED"
	EmployeeStatusTerminated EmployeeStatus = "TERMINATED"
)

type Employee struct {
	ProfileID           uuid.UUID      `db:"profile_id"`
	ClinicID            uuid.UUID      `db:"clinic_id"`
	JobTitle            *string        `db:"job_title"`
	EducationLevel      *string        `db:"education_level"`
	EmploymentStartDate *time.Time     `db:"employment_start_date"`
	PasswordHash        *string        `db:"password_hash"`
	Status              EmployeeStatus `db:"status"`
	LastLoginAt         *time.Time     `db:"last_login_at"`
	InvitedByID         *uuid.UUID     `db:"invited_by"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	Profile             Profile        `db:"-"` // Loaded separately
	Roles               []Role         `db:"-"` // Loaded separately
}

func (e *Employee) ToAuthPayload(duration time.Duration) (*security.AuthPayload, error) {
	roleIDs := make([]uuid.UUID, len(e.Roles))
	permissionSet := make(map[string]struct{})
	for i, role := range e.Roles {
		roleIDs[i] = role.ID
		for _, p := range role.Permissions {
			permissionSet[p.PermissionKey] = struct{}{}
		}
	}

	permissions := make([]string, 0, len(permissionSet))
	for p := range permissionSet {
		permissions = append(permissions, p)
	}

	return security.NewAuthPayload(e.ProfileID, e.ClinicID, roleIDs, permissions, duration)
}
