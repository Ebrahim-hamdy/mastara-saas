// Package patient contains the concrete implementation of the patient service.
package patient

import (
	"context"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/model"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/shared/database"
	"github.com/Ebrahim-hamdy/mastara-saas/internal/shared/service"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultService is the concrete implementation of the patient.Service interface.
type defaultService struct {
	service.BaseService
	repo Repository
	db   *pgxpool.Pool
}

// NewService creates a new instance of the patient service.
func NewService(txManager database.TxManager, repo Repository, db *pgxpool.Pool) Service {
	return &defaultService{
		BaseService: service.BaseService{Tx: txManager},
		repo:        repo,
		db:          db,
	}
}

// FindOrCreateGuest orchestrates the "Smart Upsert" logic for guest bookings.
func (s *defaultService) FindOrCreateGuestForBooking(ctx context.Context, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error) {
	var profile *model.Profile
	err := s.RunInTransaction(ctx, func(tx pgx.Tx) error {
		p, err := s.repo.FindOrCreateGuestForBooking(ctx, tx, clinicID, fullName, phoneNumber)
		if err != nil {
			return err
		}
		profile = p
		return nil
	})
	return profile, err
}

// RegisterNewPatient handles the creation of a fully-detailed patient profile by staff.
func (s *defaultService) RegisterNewPatient(ctx context.Context, clinicID uuid.UUID, req RegisterPatientRequest) (*model.Profile, error) {
	var profile *model.Profile
	err := s.RunInTransaction(ctx, func(tx pgx.Tx) error {
		existing, err := s.repo.FindOrCreateGuestForBooking(ctx, tx, clinicID, req.FullName, req.PhoneNumber)
		if err != nil {
			return fmt.Errorf("failed during profile lookup: %w", err)
		}

		// If the profile is already fully registered, this is a conflict.
		if existing.ProfileStatus == model.ProfileStatusRegistered {
			return apierror.NewBadRequest("A registered patient with this phone number already exists.", nil)
		}
		profile.ProfileStatus = model.ProfileStatusRegistered

		updatedProfile, updateErr := s.upsertProfile(ctx, tx, existing, req)

		profile = updatedProfile
		return updateErr
	})

	return profile, err

}

// CompleteGuestRegistration transitions a guest profile to a registered state.
func (s *defaultService) CompleteGuestRegistration(ctx context.Context, clinicID uuid.UUID, req CompleteGuestRequest) (*model.Profile, error) {
	var profile *model.Profile
	err := s.RunInTransaction(ctx, func(tx pgx.Tx) error {
		existing, err := s.repo.FindByID(ctx, s.db, req.ClinicID, req.ProfileID)
		if err != nil {
			return err
		}

		// If a guest is being updated, they become registered.
		if profile.ProfileStatus == model.ProfileStatusGuest {
			profile.ProfileStatus = model.ProfileStatusRegistered
		}

		updatedProfile, updateErr := s.upsertProfile(ctx, tx, existing, req)

		profile = updatedProfile
		return updateErr

	})

	return profile, err
}

// GetProfileByID retrieves a single patient profile.
func (s *defaultService) GetProfileByID(ctx context.Context, clinicID, profileID uuid.UUID) (*model.Profile, error) {
	profile, err := s.repo.FindByID(ctx, s.db, clinicID, profileID)
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
	return s.repo.List(ctx, s.db, clinicID, offset, pageSize)
}

func (s *defaultService) upsertProfile(ctx context.Context, tx pgx.Tx, profile *model.Profile, req ProfileUpdater) (*model.Profile, error) {
	profile.FullName = req.GetFullName()
	profile.Email = req.GetEmail()
	profile.NationalID = req.GetNationalID()
	profile.DateOfBirth = req.GetDateOfBirth()

	// The calling method is responsible for setting the correct status.
	if err := s.repo.Update(ctx, tx, profile); err != nil {
		return nil, apierror.NewInternalServer(fmt.Errorf("failed to update profile: %w", err))
	}
	return profile, nil
}
