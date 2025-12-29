package iam

import (
	"context"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/gofrs/uuid"
)

// Repository defines the contract for data access operations for the IAM module.
type Repository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindUserByEmail(ctx context.Context, clinicID uuid.UUID, email string) (*model.User, error)
	FindUserByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*model.User, error)
	FindUserByID(ctx context.Context, clinicID uuid.UUID, id uuid.UUID) (*model.User, error)
	FindRolesForUser(ctx context.Context, userID uuid.UUID) ([]model.Role, error)
}
