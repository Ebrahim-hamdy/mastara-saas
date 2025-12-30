// Package patient contains the concrete implementation of the patient service.
package patient

import (
	"context"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gofrs/uuid"
)

// defaultService is the concrete implementation of the patient.Service interface.
type defaultService struct {
	repo Repository
}

// NewService creates a new instance of the patient service.
func NewService(repo Repository) Service {
	return &defaultService{repo: repo}
}

// FindOrCreateGuest orchestrates the "Smart Upsert" logic for guest bookings.
func (s *defaultService) FindOrCreateGuestForBooking(ctx context.Context, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error) {
	if fullName == "" || phoneNumber == "" {
		return nil, apierror.NewBadRequest("Full name and phone number are required for guest booking.", nil)
	}
	// This call is atomic and race-condition-safe thanks to the repository's implementation.
	profile, err := s.repo.FindOrCreateGuestForBooking(ctx, clinicID, fullName, phoneNumber)
	if err != nil {
		return nil, apierror.NewInternalServer(fmt.Errorf("could not find or create guest profile: %w", err))
	}
	return profile, nil
}

// RegisterNewPatient handles the creation of a fully-detailed patient profile by staff.
func (s *defaultService) RegisterNewPatient(ctx context.Context, req RegisterPatientRequest) (*model.Profile, error) {
	// This is now much cleaner. We create a guest record first to ensure atomicity and then update it.
	profile, err := s.repo.FindOrCreateGuestForBooking(ctx, req.ClinicID, req.FullName, req.PhoneNumber)
	if err != nil {
		return nil, apierror.NewInternalServer(fmt.Errorf("failed during initial profile creation: %w", err))
	}

	if profile.ProfileStatus == model.ProfileStatusRegistered {
		return nil, apierror.NewBadRequest("A registered patient with this phone number already exists.", nil)
	}

	profile.ProfileStatus = model.ProfileStatusRegistered
	return s.upsertProfile(ctx, profile, req)

}

// CompleteGuestRegistration transitions a guest profile to a registered state.
func (s *defaultService) CompleteGuestRegistration(ctx context.Context, req CompleteGuestRequest) (*model.Profile, error) {
	profile, err := s.repo.FindByID(ctx, req.ClinicID, req.ProfileID)
	if err != nil {
		return nil, err // Repository returns a well-typed NotFound error
	}

	// If a guest is being updated, they become registered.
	if profile.ProfileStatus == model.ProfileStatusGuest {
		profile.ProfileStatus = model.ProfileStatusRegistered
	}

	return s.upsertProfile(ctx, profile, req)
}

// GetProfileByID retrieves a single patient profile.
func (s *defaultService) GetProfileByID(ctx context.Context, clinicID, profileID uuid.UUID) (*model.Profile, error) {
	profile, err := s.repo.FindByID(ctx, clinicID, profileID)
	if err != nil {
		// The repository already returns a correctly typed apierror.NotFound
		return nil, err
	}
	return profile, nil
}

func (s *defaultService) ListProfiles(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]model.Profile, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, clinicID, offset, pageSize)
}

func (s *defaultService) upsertProfile(ctx context.Context, profile *model.Profile, req ProfileUpdater) (*model.Profile, error) {
	profile.FullName = req.GetFullName()
	profile.Email = req.GetEmail()
	profile.NationalID = req.GetNationalID()
	profile.DateOfBirth = req.GetDateOfBirth()

	// The calling method is responsible for setting the correct status.
	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, apierror.NewInternalServer(fmt.Errorf("failed to update profile: %w", err))
	}
	return profile, nil
}
