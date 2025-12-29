package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/config"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/infra/security"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/middleware"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/store" // Import the store package
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gofrs/uuid"
)

// serviceImpl is the concrete implementation of the iam.Service interface.
type serviceImpl struct {
	repo   Repository
	sec    *security.PasetoManager
	config *config.Config
}

// NewService creates a new instance of the IAM service.
func NewService(repo Repository, sec *security.PasetoManager, config *config.Config) Service {
	return &serviceImpl{
		repo:   repo,
		sec:    sec,
		config: config,
	}
}

// RegisterUser handles the business logic for creating a new user.
// It hashes the password and persists the new user to the database.
func (s *serviceImpl) RegisterUser(ctx context.Context, req RegisterUserRequest) (*model.User, error) {
	// 1. Extract inviter's context for auditing and tenancy.
	payload := middleware.GetAuthPayload(ctx)
	if payload == nil {
		return nil, apierror.NewInternalServer(errors.New("auth payload not found in context for registration"))
	}

	// 2. Hash the password.
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, apierror.NewInternalServer(fmt.Errorf("failed to hash password: %w", err))
	}

	// 3. Construct the domain model.
	newUser := &model.User{
		ID:           uuid.Must(uuid.NewV7()),
		ClinicID:     payload.ClinicID,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: &hashedPassword,
		FullName:     req.FullName,
		JobTitle:     req.JobTitle,
		Status:       "ACTIVE", // Or "INVITED" if we build an invite flow
		InvitedByID:  &payload.UserID,
	}

	// 4. Persist to the database with proper error handling.
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		// --- THIS IS THE CRITICAL CORRECTION ---
		if store.IsUniqueViolationError(err) {
			// This is a predictable business rule violation, not a server error.
			return nil, apierror.NewBadRequest("A user with this email or phone number already exists.", err)
		}
		// For all other database errors, it's an internal server issue.
		return nil, apierror.NewInternalServer(fmt.Errorf("failed to create user in db: %w", err))
		// --- END CORRECTION ---
	}

	return newUser, nil
}

// LoginUser handles the business logic for authenticating a user and issuing a token.
func (s *serviceImpl) LoginUser(ctx context.Context, req LoginUserRequest) (string, *model.User, error) {
	// 1. Extract clinic_id from context for tenant-scoped lookup.
	payload := middleware.GetAuthPayload(ctx)
	if payload == nil {
		return "", nil, apierror.NewInternalServer(errors.New("auth payload not found in context for login"))
	}

	// 2. Find the user by email or phone.
	var user *model.User
	var err error
	if req.Email != nil {
		user, err = s.repo.FindUserByEmail(ctx, payload.ClinicID, *req.Email)
	} else if req.Phone != nil {
		user, err = s.repo.FindUserByPhone(ctx, payload.ClinicID, *req.Phone)
	} else {
		return "", nil, apierror.NewBadRequest("email or phone is required for login", nil)
	}

	if err != nil {
		if _, ok := err.(*apierror.APIError); ok {
			return "", nil, apierror.NewUnauthorized("invalid credentials", err)
		}
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to find user: %w", err))
	}

	// 3. Verify the password.
	if user.PasswordHash == nil {
		return "", nil, apierror.NewUnauthorized("invalid credentials (user has no password set)", nil)
	}
	if err := security.ComparePasswordAndHash(req.Password, *user.PasswordHash); err != nil {
		return "", nil, err
	}

	// 4. Fetch user's roles and permissions.
	roles, err := s.repo.FindRolesForUser(ctx, user.ID)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to fetch user roles: %w", err))
	}
	user.Roles = roles

	// 5. Generate the authentication token.
	authPayload, err := user.ToAuthPayload(s.config.Security.TokenDuration)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to create auth payload: %w", err))
	}

	token, err := s.sec.CreateToken(authPayload)
	if err != nil {
		return "", nil, apierror.NewInternalServer(fmt.Errorf("failed to create token: %w", err))
	}

	return token, user, nil
}
