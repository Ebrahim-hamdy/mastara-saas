-- This migration establishes the core tables for Identity and Access Management (IAM),
-- including a robust users table, a flexible RBAC system, and the necessary
-- featuring a flexible identity model (email or phone number) and a robust RBAC system.
-- constraints for a secure, auditable, multi-tenant environment.

-- The uuid_generate_v7() function and the trigger_set_timestamp() function are assumed
-- to exist from the previous migration (000001).

-- Create a custom ENUM type for user status to enforce data integrity.
CREATE TYPE user_status AS ENUM (
    'INVITED',    -- Account created by an admin, user has not set a password or logged in.
    'ACTIVE',     -- User has completed onboarding and can use the system.
    'SUSPENDED',  -- Temporarily disabled by an admin, can be reactivated.
    'DEACTIVATED' -- Permanently disabled, cannot be reactivated.
);

CREATE TABLE permissions (
    id SMALLINT PRIMARY KEY,
    permission_key VARCHAR(100) NOT NULL UNIQUE
);
COMMENT ON TABLE permissions IS 'Defines atomic, system-wide permissions (e.g., "patient.create", "finance.view").';

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_roles_clinic_name UNIQUE (clinic_id, name),
    CONSTRAINT chk_system_role_clinic_id CHECK ( (is_system_role AND clinic_id IS NULL) OR (NOT is_system_role) )
);
COMMENT ON TABLE roles IS 'Groups permissions. Can be system-wide (clinic_id IS NULL) or clinic-specific.';

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id SMALLINT NOT NULL REFERENCES permissions(id) ON DELETE RESTRICT,
    PRIMARY KEY (role_id, permission_id)
);
COMMENT ON TABLE role_permissions IS 'Many-to-many relationship between roles and permissions.';

-- Users are the staff members who log into the system.
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,

    -- Flexible Identity: At least one of email or phone_number must be provided.
    email VARCHAR(255),
    phone_number VARCHAR(50),

    -- Credentials
    password_hash VARCHAR(255), -- Can be NULL for 'INVITED' status until user sets it.

    -- Profile Information
    full_name VARCHAR(255) NOT NULL,
    job_title VARCHAR(100),      -- e.g., 'Dentist', 'Receptionist'

    -- Status and Lifecycle Management
    status user_status NOT NULL DEFAULT 'INVITED',
    last_login_at TIMESTAMPTZ,

    -- Auditing and Onboarding Trail
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL, -- Who created this user? Self-referencing FK.
    
    -- Standard Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ, -- For soft deletes

    -- === CRITICAL DATA INTEGRITY CONSTRAINTS ===
    -- 1. Enforce that at least one contact method (email or phone) exists.
    CONSTRAINT chk_user_contact_method CHECK (email IS NOT NULL OR phone_number IS NOT NULL),

    -- 2. Enforce that an active user must have a password.
    CONSTRAINT chk_password_for_active_user CHECK ( (status != 'INVITED' AND password_hash IS NOT NULL) OR (status = 'INVITED') )
);
COMMENT ON TABLE users IS 'Represents staff members. Identity is based on either email or phone number within a clinic.';

-- === ADVANCED UNIQUENESS CONSTRAINTS ===
-- Use partial unique indexes to enforce uniqueness only on non-NULL values.
-- This is the correct way to handle optional unique fields.

-- A user's email, if provided, must be unique within their clinic.
CREATE UNIQUE INDEX idx_users_unique_email_per_clinic ON users (clinic_id, email) WHERE email IS NOT NULL;

-- A user's phone number, if provided, must be unique within their clinic.
CREATE UNIQUE INDEX idx_users_unique_phone_per_clinic ON users (clinic_id, phone_number) WHERE phone_number IS NOT NULL;

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and their assigned roles.';

-- Apply the timestamp trigger to all mutable tables.
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON roles
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Seed the initial, non-negotiable permissions.
INSERT INTO permissions (id, permission_key) VALUES
(1, 'users.invite'), (2, 'users.read'), (3, 'users.update'), (4, 'users.deactivate'),
(10, 'patients.create'), (11, 'patients.read'), (12, 'patients.update'), (13, 'patients.delete'),
(20, 'appointments.create'), (21, 'appointments.read'), (22, 'appointments.update'), (23, 'appointments.delete'),
(30, 'finance.invoice.create'), (31, 'finance.invoice.read'), (32, 'finance.payment.record'), (33, 'finance.reports.view'),
(40, 'roles.create'), (41, 'roles.read'), (42, 'roles.update'), (43, 'roles.delete');