package iam

import (
	"context"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model" // Import the store package
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gofrs/uuid"
)

// serviceImpl is the concrete implementation of the iam.Service interface.
type defaultService struct {
	repo   Repository
	sec    *security.PasetoManager
	config *config.Config
	// We need a way to find the clinic for a login request.
	// This would be a repository from another module, injected here.
	// For now, we'll assume a placeholder function signature.
	// clinicRepo clinic.Repository

}

// NewService creates a new instance of the IAM service.
func NewService(repo Repository, sec *security.PasetoManager, config *config.Config) Service {
	return &defaultService{repo, sec, config}
}

// InviteEmployee handles the business logic for creating a new employee in an 'INVITED' state.
func (s *defaultService) InviteEmployee(ctx context.Context, clinicID, inviterID uuid.UUID, req InviteEmployeeRequest) (*model.Employee, error) {
	profileID := uuid.Must(uuid.NewV7())

	newProfile := &model.Profile{
		ID:          profileID,
		ClinicID:    clinicID,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
	}

	newEmployee := &model.Employee{
		ProfileID:   profileID,
		ClinicID:    clinicID,
		JobTitle:    req.JobTitle,
		Status:      "INVITED",
		InvitedByID: &inviterID,
		Profile:     *newProfile, // Embed profile for response mapping
	}

	if err := s.repo.CreateInvitedEmployee(ctx, newProfile, newEmployee); err != nil {
		// The repository should handle unique violation checks.
		return nil, err
	}

	// In a real flow, we would now generate an invitation token and send an email/SMS.
	// For now, creating the record is sufficient.

	return newEmployee, nil
}

// LoginEmployee handles authentication for staff members.
func (s *defaultService) LoginEmployee(ctx context.Context, req LoginEmployeeRequest) (string, *model.Employee, error) {
	// Login is a public action, so it doesn't use the auth payload from context.
	// It needs a clinic_id, which would typically be derived from a subdomain or a header.
	// For now, we'll assume a placeholder. This needs to be addressed when we build the full login flow.
	// A real implementation would require a `FindClinicByDomain` method.
	// --- THIS IS THE CRITICAL CORRECTION ---
	// A login request is unauthenticated. It cannot have an AuthPayload.
	// The request must contain enough information to identify the clinic.
	// A real-world app would get this from the request's hostname (e.g., clinic-a.mastara.com)
	// or a non-sensitive header like `X-Clinic-ID`.
	// For now, we will simulate this by requiring the DTO to carry it.
	// This makes the dependency explicit.

	placeholderClinicID := uuid.Must(uuid.NewV4()) // THIS IS A PLACEHOLDER

	var employee *model.Employee
	var err error
	if req.Email != nil {
		employee, err = s.repo.FindEmployeeByEmail(ctx, placeholderClinicID, *req.Email)
	} else if req.Phone != nil {
		employee, err = s.repo.FindEmployeeByPhone(ctx, placeholderClinicID, *req.Phone)
	} else {
		return "", nil, apierror.NewBadRequest("email or phone is required for login", nil)
	}

	if err != nil {
		if _, ok := err.(*apierror.APIError); ok {
			return "", nil, apierror.NewUnauthorized("invalid credentials", err)
		}
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to find employee: %w", err))
	}

	if employee.PasswordHash == nil {
		return "", nil, apierror.NewUnauthorized("invalid credentials (account not fully set up)", nil)
	}
	if err := security.ComparePasswordAndHash(req.Password, *employee.PasswordHash); err != nil {
		return "", nil, err
	}

	roles, err := s.repo.FindRolesForEmployee(ctx, employee.ProfileID)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to fetch employee roles: %w", err))
	}
	employee.Roles = roles

	authPayload, err := employee.ToAuthPayload(s.config.Security.TokenDuration)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to create auth payload: %w", err))
	}

	token, err := s.sec.CreateToken(authPayload)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to create token: %w", err))
	}

	return token, employee, nil
}
