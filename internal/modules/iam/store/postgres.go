package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier defines the common methods between pgx.Tx and *pgxpool.Pool.
// This allows repository methods to be used both within and outside of transactions.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// pgxRepository is the PostgreSQL implementation of the iam.Repository.
type pgxRepository struct {
	db *pgxpool.Pool
}

// NewPgxRepository creates a new instance of the IAM repository.
func NewPgxRepository(db *pgxpool.Pool) *pgxRepository {
	return &pgxRepository{db: db}
}

// isUniqueViolationError checks if a given error is a PostgreSQL unique constraint violation (code 23505).
func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// FindOrCreateGuest atomically inserts a guest or retrieves the existing one.
// It relies on a "Writeable CTE with UNION" pattern to be efficient and thread-safe.
// This executes exactly ONE round trip to the DB and optimizes index usage.
func (r *pgxRepository) FindOrCreateGuest(ctx context.Context, querier Querier, clinicID uuid.UUID, fullName string, phoneNumber string) (*model.Profile, error) {
	profile := &model.Profile{}

	// Generating ID here ensures it matches the INSERT arg if new
	newID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("profile_repo: failed to generate uuidv7: %w", err)
	}

	// 2. SQL QUERY (Hybrid Upsert Pattern)
	// - Tries to Insert New ID.
	// - If conflict, finds existing ID.
	// - UNION ALL merges the logic into one set.
	query := `
        WITH new_row AS (
            INSERT INTO profiles (id, clinic_id, full_name, phone_number, profile_status)
            VALUES ($1, $2, $3, $4, 'GUEST')
            ON CONFLICT (clinic_id, phone_number) WHERE deleted_at IS NULL DO NOTHING
            RETURNING id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data, created_at, updated_at, deleted_at
        )
        SELECT * FROM new_row
        UNION ALL
        SELECT id, clinic_id, full_name, phone_number, email, national_id, date_of_birth, profile_status, extended_data, created_at, updated_at, deleted_at
        FROM profiles
        WHERE clinic_id = $2 AND phone_number = $4 AND deleted_at IS NULL
        LIMIT 1;
    `

	// 3. EXECUTION
	// Note: profile.Email and profile.NationalID MUST be pointers (*string)
	// in your model definition to handle the NULLs safely, or scanner will panic.
	err = querier.QueryRow(ctx, query, newID, clinicID, fullName, phoneNumber).Scan(
		&profile.ID,
		&profile.ClinicID,
		&profile.FullName,
		&profile.PhoneNumber,
		&profile.Email,      // Caution: Scan handles DB NULLs if this is *string
		&profile.NationalID, // Caution: Scan handles DB NULLs if this is *string
		&profile.DateOfBirth,
		&profile.ProfileStatus,
		&profile.ExtendedData,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Should be impossible due to UPSERT logic, but catch just in case.
			return nil, fmt.Errorf("profile_repo: critical upsert failure for guest: %w", err)
		}
		return nil, fmt.Errorf("profile_repo: query execution failed: %w", err)
	}

	return profile, nil

}

// CreateUser inserts a new user record into the database.
// CreateInvitedEmployee creates a profile and an employee record within a single transaction.
func (r *pgxRepository) CreateInvitedEmployee(ctx context.Context, tx pgx.Tx, profile *model.Profile, employee *model.Employee) error {
	profileQuery := `
        INSERT INTO profiles (id, clinic_id, full_name, email, phone_number, profile_status)
        VALUES ($1, $2, $3, $4, $5, 'REGISTERED')`
	if _, err := tx.Exec(ctx, profileQuery, profile.ID, profile.ClinicID, profile.FullName, profile.Email, profile.PhoneNumber); err != nil {
		if IsUniqueViolationError(err) {
			return apierror.NewBadRequest("A profile with this email or phone number already exists.", err)
		}
		return fmt.Errorf("store.CreateInvitedEmployee: failed to insert profile: %w", err)
	}

	employeeQuery := `
        INSERT INTO employees (profile_id, clinic_id, job_title, status, invited_by)
        VALUES ($1, $2, $3, $4, $5)`
	if _, err := tx.Exec(ctx, employeeQuery, employee.ProfileID, employee.ClinicID, employee.JobTitle, employee.Status, employee.InvitedByID); err != nil {
		return fmt.Errorf("store.CreateInvitedEmployee: failed to insert employee: %w", err)
	}

	return nil
}

// FindEmployeeByEmail finds a user by their email within the specified clinic.
func (r *pgxRepository) FindEmployeeByEmail(ctx context.Context, clinicID uuid.UUID, email string) (*model.Employee, error) {
	employee := &model.Employee{}
	query := `
        SELECT id, clinic_id, email, phone_number, password_hash, full_name, job_title, status, last_login_at, invited_by, created_at, updated_at, deleted_at
        FROM users
        WHERE clinic_id = $1 AND email = $2 AND deleted_at IS NULL
    `
	err := r.db.QueryRow(ctx, query, clinicID, email).Scan(
		&employee.ProfileID, &employee.ClinicID, &employee.Profile.Email, &employee.Profile.PhoneNumber, &employee.PasswordHash,
		&employee.Profile.FullName, &employee.JobTitle, &employee.Status, &employee.LastLoginAt, &employee.InvitedByID,
		&employee.CreatedAt, &employee.UpdatedAt, &employee.Profile.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFound("user", err)
		}
		return nil, fmt.Errorf("store.FindUserByEmail: failed to query user: %w", err)
	}
	return employee, nil
}

// FindEmployeeByPhone finds a user by their phone number within the specified clinic.
func (r *pgxRepository) FindEmployeeByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*model.Employee, error) {
	employee := &model.Employee{}
	query := `
        SELECT id, clinic_id, email, phone_number, password_hash, full_name, job_title, status, last_login_at, invited_by, created_at, updated_at, deleted_at
        FROM users
        WHERE clinic_id = $1 AND phone_number = $2 AND deleted_at IS NULL
    `
	err := r.db.QueryRow(ctx, query, clinicID, phone).Scan(
		&employee.ProfileID, &employee.ClinicID, &employee.Profile.Email, &employee.Profile.PhoneNumber, &employee.PasswordHash,
		&employee.Profile.FullName, &employee.JobTitle, &employee.Status, &employee.LastLoginAt, &employee.InvitedByID,
		&employee.CreatedAt, &employee.UpdatedAt, &employee.Profile.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFound("user", err)
		}
		return nil, fmt.Errorf("store.FindUserByPhone: failed to query user: %w", err)
	}
	return employee, nil
}

// FindEmployeeByIDWithDetails finds a user by their ID within the specified clinic.
func (r *pgxRepository) FindEmployeeByIDWithDetails(ctx context.Context, clinicID uuid.UUID, id uuid.UUID) (*model.Employee, error) {
	employee := &model.Employee{}
	query := `
        SELECT id, clinic_id, email, phone_number, password_hash, full_name, job_title, status, last_login_at, invited_by, created_at, updated_at, deleted_at
        FROM users
        WHERE clinic_id = $1 AND id = $2 AND deleted_at IS NULL
    `
	err := r.db.QueryRow(ctx, query, clinicID, id).Scan(
		&employee.ProfileID, &employee.ClinicID, &employee.Profile.Email, &employee.Profile.PhoneNumber, &employee.PasswordHash,
		&employee.Profile.FullName, &employee.JobTitle, &employee.Status, &employee.LastLoginAt, &employee.InvitedByID,
		&employee.CreatedAt, &employee.UpdatedAt, &employee.Profile.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierror.NewNotFound("user", err)
		}
		return nil, fmt.Errorf("store.FindUserByID: failed to query user: %w", err)
	}
	return employee, nil
}

// FindRolesForEmployee retrieves all roles (and their permissions) assigned to a user.
func (r *pgxRepository) FindRolesForEmployee(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
	query := `
        SELECT r.id, r.clinic_id, r.name, r.description, r.is_system_role,
               p.id, p.permission_key
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        LEFT JOIN role_permissions rp ON r.id = rp.role_id
        LEFT JOIN permissions p ON rp.permission_id = p.id
        WHERE ur.user_id = $1
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("store.FindRolesForUser: failed to query roles: %w", err)
	}
	defer rows.Close()

	roleMap := make(map[uuid.UUID]*model.Role)
	for rows.Next() {
		var role model.Role
		var pID sql.NullInt16
		var pKey sql.NullString

		if err := rows.Scan(&role.ID, &role.ClinicID, &role.Name, &role.Description, &role.IsSystemRole, &pID, &pKey); err != nil {
			return nil, fmt.Errorf("store.FindRolesForUser: failed to scan row: %w", err)
		}

		if _, ok := roleMap[role.ID]; !ok {
			role.Permissions = []model.Permission{}
			roleMap[role.ID] = &role
		}

		if pID.Valid && pKey.Valid {
			permission := model.Permission{ID: int16(pID.Int16), PermissionKey: pKey.String}
			roleMap[role.ID].Permissions = append(roleMap[role.ID].Permissions, permission)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.FindRolesForUser: error iterating rows: %w", err)
	}

	roles := make([]model.Role, 0, len(roleMap))
	for _, role := range roleMap {
		roles = append(roles, *role)
	}

	return roles, nil
}
