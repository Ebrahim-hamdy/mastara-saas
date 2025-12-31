// Package iam contains all the business logic for the Identity and Access Management module.
package iam

import (
	"context"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service defines the contract for the IAM module's business logic (for employees).
type Service interface {
	InviteEmployee(ctx context.Context, clinicID, inviterID uuid.UUID, req InviteEmployeeRequest) (*model.Employee, error)
	LoginEmployee(ctx context.Context, req LoginEmployeeRequest) (token string, employee *model.Employee, err error)
	// We will add AcceptInvite and other methods later.
}

// Repository defines the data access contract for employees.
type Repository interface {
	// Creates the profile and employee records in a single transaction.
	CreateInvitedEmployee(ctx context.Context, tx pgx.Tx, profile *model.Profile, employee *model.Employee) error
	FindEmployeeByEmail(ctx context.Context, clinicID uuid.UUID, email string) (*model.Employee, error)
	FindEmployeeByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*model.Employee, error)
	FindEmployeeByIDWithDetails(ctx context.Context, clinicID, profileID uuid.UUID) (*model.Employee, error)
	FindRolesForEmployee(ctx context.Context, employeeProfileID uuid.UUID) ([]model.Role, error)
}

// InviteEmployeeRequest contains the data needed to invite a new staff member.
type InviteEmployeeRequest struct {
	FullName    string
	Email       *string
	PhoneNumber *string
	JobTitle    *string
}

// LoginEmployeeRequest contains credentials for an employee login.
type LoginEmployeeRequest struct {
	ClinicID uuid.UUID // This must be provided by the handler.
	Email    *string
	Phone    *string
	Password string
}
