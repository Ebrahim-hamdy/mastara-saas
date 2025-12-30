package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ebrahim-hamdy/mastara-saas/internal/modules/iam/model"
	"github.com/Ebrahim-hamdy/mastara-saas/pkg/apierror"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// CreateUser inserts a new user record into the database.
func (r *pgxRepository) CreateUser(ctx context.Context, user *model.Employee) error {
	query := `
        INSERT INTO users (id, clinic_id, email, phone_number, password_hash, full_name, job_title, status, invited_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
	_, err := r.db.Exec(ctx, query,
		user.ProfileID, user.ClinicID, user.Profile.Email, user.Profile.PhoneNumber, user.PasswordHash,
		user.Profile.FullName, user.JobTitle, user.Status, user.InvitedByID,
	)
	if err != nil {
		return fmt.Errorf("store.CreateUser: failed to execute query: %w", err)
	}
	return nil
}

// FindUserByEmail finds a user by their email within the specified clinic.
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

// FindUserByPhone finds a user by their phone number within the specified clinic.
func (r *pgxRepository) FindUserByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*model.Employee, error) {
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

// FindUserByID finds a user by their ID within the specified clinic.
func (r *pgxRepository) FindUserByID(ctx context.Context, clinicID uuid.UUID, id uuid.UUID) (*model.Employee, error) {
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

// FindRolesForUser retrieves all roles (and their permissions) assigned to a user.
func (r *pgxRepository) FindRolesForUser(ctx context.Context, userID uuid.UUID) ([]model.Role, error) {
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
