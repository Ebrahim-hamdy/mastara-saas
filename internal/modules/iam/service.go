// Package iam contains all the business logic for the Identity and Access Management module.
package iam

import (
	"context"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
)

// Service defines the contract for the IAM module's business logic.
type Service interface {
	RegisterUser(ctx context.Context, req RegisterUserRequest) (*model.User, error)
	LoginUser(ctx context.Context, req LoginUserRequest) (token string, user *model.User, err error)
}

// RegisterUserRequest contains the data needed to register a new user.
type RegisterUserRequest struct {
	FullName    string
	Email       *string
	PhoneNumber *string
	Password    string
	JobTitle    *string
}

// LoginUserRequest contains the credentials for a user login attempt.
type LoginUserRequest struct {
	Email    *string
	Phone    *string
	Password string
}
