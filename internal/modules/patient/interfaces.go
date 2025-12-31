// Package patient contains all business logic for patient management.
package patient

import (
	"context"
	"time"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/model"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/store"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/shared/database"
	"github.com/google/uuid"
)

// Querier is an alias for the store's Querier interface.
type Querier interface {
	store.Querier
}

// Service defines the contract for the Patient module's business logic.
type Service interface {
	// RegisterNewPatient is used by staff to create a fully-registered patient profile at once.
	RegisterNewPatient(ctx context.Context, clinicID uuid.UUID, req RegisterPatientRequest) (*model.Profile, error)

	// CompleteGuestRegistration is used by staff to enrich a guest profile with full details.
	CompleteGuestRegistration(ctx context.Context, clinicID uuid.UUID, req CompleteGuestRequest) (*model.Profile, error)

	// UpdatePatientDetails(ctx context.Context, req CompleteGuestRequest) (*model.Profile, error)

	// GetProfileByID retrieves a single patient profile.
	GetProfileByID(ctx context.Context, clinicID, profileID uuid.UUID) (*model.Profile, error)

	ListProfiles(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]model.Profile, error)

	// Public/Guest-facing methods
	FindOrCreateGuestForBooking(ctx context.Context, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error)
}

// Repository defines the contract for data access operations for the Patient/Profile module.
type Repository interface {
	// FindOrCreateGuest atomically finds a profile by phone number or creates a new one if it doesn't exist.
	// This is the core of the "Guest Checkout" booking flow.
	FindOrCreateGuestForBooking(ctx context.Context, querier database.Querier, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error)

	FindByID(ctx context.Context, querier database.Querier, clinicID, profileID uuid.UUID) (*model.Profile, error)
	Create(ctx context.Context, querier database.Querier, profile *model.Profile) error
	Update(ctx context.Context, querier database.Querier, profile *model.Profile) error
	List(ctx context.Context, querier database.Querier, clinicID uuid.UUID, offset, limit int) ([]model.Profile, error)
}

// RegisterPatientRequest contains all data for creating a new, fully registered patient.
type RegisterPatientRequest struct {
	ClinicID    uuid.UUID
	FullName    string
	PhoneNumber string
	Email       *string
	NationalID  *string
	DateOfBirth *time.Time
}

func (r RegisterPatientRequest) GetFullName() string        { return r.FullName }
func (r RegisterPatientRequest) GetEmail() *string          { return r.Email }
func (r RegisterPatientRequest) GetNationalID() *string     { return r.NationalID }
func (r RegisterPatientRequest) GetDateOfBirth() *time.Time { return r.DateOfBirth }

// CompleteGuestRequest contains the data to upgrade a guest profile to a registered one.
type CompleteGuestRequest struct {
	ClinicID    uuid.UUID
	ProfileID   uuid.UUID
	FullName    string
	Email       *string
	NationalID  *string
	DateOfBirth *time.Time
}

func (r CompleteGuestRequest) GetFullName() string        { return r.FullName }
func (r CompleteGuestRequest) GetEmail() *string          { return r.Email }
func (r CompleteGuestRequest) GetNationalID() *string     { return r.NationalID }
func (r CompleteGuestRequest) GetDateOfBirth() *time.Time { return r.DateOfBirth }

// ProfileUpdater is an interface that both Register and Update requests will satisfy.
// This allows for a single, DRY upsert method in the service.
type ProfileUpdater interface {
	GetFullName() string
	GetEmail() *string
	GetNationalID() *string
	GetDateOfBirth() *time.Time
}
