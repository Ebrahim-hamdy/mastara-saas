// Package store provides the database implementation for the patient/profile repository.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/patient/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxProfileRepository is the PostgreSQL implementation of the patient.Repository.
type pgxProfileRepository struct {
	db *pgxpool.Pool
}

// NewPgxProfileRepository creates a new instance of the profile repository.
func NewPgxProfileRepository(db *pgxpool.Pool) *pgxProfileRepository {
	return &pgxProfileRepository{db: db}
}

// Create inserts a new profile record into the database.
func (r *pgxProfileRepository) Create(ctx context.Context, profile *model.Profile) error {
	query := `
        INSERT INTO profiles (id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.Exec(ctx, query,
		profile.ID, profile.ClinicID, profile.FullName, profile.PhoneNumber, profile.Email,
		profile.NationalID, profile.DateOfBirth, profile.ProfileStatus, profile.ExtendedData,
	)
	if err != nil {
		// Check for unique constraint violation on phone or email
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apierror.NewBadRequest("A patient with this phone number or email already exists in this clinic.", err)
		}
		return fmt.Errorf("store.Create: failed to execute query: %w", err)
	}
	return nil
}

// FindOrCreateGuest atomically finds a profile by phone number for a given clinic,
// or creates a new 'GUEST' profile if one does not exist. This is implemented
// using a CTE with ON CONFLICT to ensure it is a single, race-condition-safe operation.
func (r *pgxProfileRepository) FindOrCreateGuestForBooking(ctx context.Context, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error) {
	profile := &model.Profile{}

	// This query is the heart of the "Smart Upsert" logic.
	// 1. `inserted` CTE: Attempts to insert a new guest profile.
	//    `ON CONFLICT (clinic_id, phone_number) DO NOTHING` ensures that if a profile
	//    with that phone number already exists for the clinic, the insert is silently ignored.
	// 2. `SELECT`: We then select the profile that matches the phone number.
	//    - If the insert succeeded, this select will find the newly created row.
	//    - If the insert was ignored (due to conflict), this select will find the existing row.
	query := `
        WITH inserted AS (
            INSERT INTO profiles (id, clinic_id, full_name, phone_number, profile_status)
            VALUES (uuid_generate_v7(), $1, $2, $3, 'GUEST')
            ON CONFLICT (clinic_id, phone_number) DO NOTHING
            RETURNING *
        )
        SELECT id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data, created_at, updated_at, deleted_at
        FROM profiles
        WHERE clinic_id = $1 AND phone_number = $3 AND deleted_at IS NULL
    `

	err := r.db.QueryRow(ctx, query, clinicID, fullName, phoneNumber).Scan(
		&profile.ID, &profile.ClinicID, &profile.FullName, &profile.PhoneNumber, &profile.Email,
		&profile.NationalID, &profile.DateOfBirth, &profile.ProfileStatus, &profile.ExtendedData,
		&profile.CreatedAt, &profile.UpdatedAt, &profile.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// This case should be practically impossible with the CTE logic, but is included for safety.
			return nil, apierror.NewInternalServer(fmt.Errorf("failed to find or create guest profile, though this should not happen: %w", err))
		}
		return nil, fmt.Errorf("store.FindOrCreateGuest: failed to execute query: %w", err)
	}

	return profile, nil
}

// FindByID finds a profile by its ID, scoped to the given clinic.
func (r *pgxProfileRepository) FindByID(ctx context.Context, clinicID, profileID uuid.UUID) (*model.Profile, error) {
	profile := &model.Profile{}
	query := `
        SELECT id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data, created_at, updated_at, deleted_at
        FROM profiles
        WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL
    `
	err := r.db.QueryRow(ctx, query, clinicID, profileID).Scan(
		&profile.ID, &profile.ClinicID, &profile.FullName, &profile.PhoneNumber, &profile.Email,
		&profile.NationalID, &profile.DateOfBirth, &profile.ProfileStatus, &profile.ExtendedData,
		&profile.CreatedAt, &profile.UpdatedAt, &profile.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFound("profile", err)
		}
		return nil, fmt.Errorf("store.FindByID: failed to query profile: %w", err)
	}
	return profile, nil
}

// Update persists changes to a profile record.
func (r *pgxProfileRepository) Update(ctx context.Context, profile *model.Profile) error {
	query := `
        UPDATE profiles
        SET full_name = $1, phone_number = $2, email = $3, national_id = $4, date_of_birth = $5, profile_status = $6, extended_data = $7
        WHERE id = $8 AND clinic_id = $9
    `
	cmdTag, err := r.db.Exec(ctx, query,
		profile.FullName, profile.PhoneNumber, profile.Email, profile.NationalID,
		profile.DateOfBirth, profile.ProfileStatus, profile.ExtendedData,
		profile.ID, profile.ClinicID,
	)
	if err != nil {
		return fmt.Errorf("store.Update: failed to execute update: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return apierror.NewNotFound("profile", nil)
	}
	return nil
}

func (r *pgxProfileRepository) List(ctx context.Context, clinicID uuid.UUID, offset, limit int) ([]model.Profile, error) {
	var profiles []model.Profile
	query := `
        SELECT id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data, created_at, updated_at, deleted_at
        FROM profiles
        WHERE clinic_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Query(ctx, query, clinicID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("store.List: failed to query profiles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var profile model.Profile
		if err := rows.Scan(
			&profile.ID, &profile.ClinicID, &profile.FullName, &profile.PhoneNumber, &profile.Email,
			&profile.NationalID, &profile.DateOfBirth, &profile.ProfileStatus, &profile.ExtendedData,
			&profile.CreatedAt, &profile.UpdatedAt, &profile.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("store.List: failed to scan profile row: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.List: error iterating profile rows: %w", err)
	}

	return profiles, nil
}
