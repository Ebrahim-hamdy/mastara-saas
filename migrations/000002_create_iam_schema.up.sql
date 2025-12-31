-- This migration establishes the core tables for Identity and Access Management (IAM),
-- centered around a unified 'profiles' table for all individuals (staff and patients).

-- The uuid_generate_v7() function and the trigger_set_timestamp() function are assumed
-- to exist from the previous migration (000001).

CREATE TYPE profile_status AS ENUM (
    'GUEST',      -- A patient profile created via fast booking, not a system user.
    'REGISTERED', -- A patient profile that has a linked account.
    'ARCHIVED'
);

CREATE TYPE employee_status AS ENUM (
    'INVITED',
    'ACTIVE',
    'SUSPENDED',
    'TERMINATED'
);

-- The unified 'profiles' table for every person in the system.
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    
    -- Core PII
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    email VARCHAR(255),
    national_id VARCHAR(100),
    date_of_birth DATE,

    -- Patient-specific status
    profile_status profile_status NOT NULL DEFAULT 'GUEST',

    -- Flexible data storage for non-core or custom fields.
    extended_data JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- A person must have at least one contact method.
    CONSTRAINT chk_profile_contact_method CHECK (email IS NOT NULL OR phone_number IS NOT NULL)
    
    -- Partial unique indexes for optional, tenant-scoped identifiers.
    -- CONSTRAINT uq_profiles_clinic_phone UNIQUE (clinic_id, phone_number) WHERE phone_number IS NOT NULL,
    -- CONSTRAINT uq_profiles_clinic_email UNIQUE (clinic_id, email) WHERE email IS NOT NULL
);
COMMENT ON TABLE profiles IS 'Canonical store for all individuals (patients and staff). A person''s PII lives here.';

CREATE UNIQUE INDEX idx_profiles_unique_active_phone_per_clinic ON profiles (clinic_id, phone_number) WHERE phone_number IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_profiles_unique_active_email_per_clinic ON profiles (clinic_id, email) WHERE email IS NOT NULL AND deleted_at IS NULL;

-- The 'employees' table for staff-specific data. One-to-one with profiles.
CREATE TABLE employees (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,

    -- Employment-specific data
    job_title VARCHAR(100),
    education_level TEXT,
    employment_start_date DATE,
    employment_end_date DATE,

    -- Login credentials
    password_hash VARCHAR(255), -- Nullable for 'INVITED' status.

    -- Status and Lifecycle
    status employee_status NOT NULL DEFAULT 'INVITED',
    last_login_at TIMESTAMPTZ,

    -- Auditing
    invited_by UUID REFERENCES employees(profile_id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_password_for_active_employee CHECK ( (status != 'INVITED' AND password_hash IS NOT NULL) OR (status = 'INVITED') )
);
COMMENT ON TABLE employees IS 'Stores employment-specific data for staff members. Has a 1-to-1 relationship with the profiles table.';

-- RBAC Tables
CREATE TABLE permissions (
    id SMALLINT PRIMARY KEY,
    permission_key VARCHAR(100) NOT NULL UNIQUE
);
COMMENT ON TABLE permissions IS 'Defines atomic, system-wide permissions.';

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID REFERENCES clinics(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system_role BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    -- CONSTRAINT uq_roles_clinic_name UNIQUE (clinic_id, name),
    CONSTRAINT chk_system_role_clinic_id CHECK ( (is_system_role AND clinic_id IS NULL) OR (NOT is_system_role) )
);
COMMENT ON TABLE roles IS 'Groups permissions. Can be system-wide or clinic-specific.';

-- HARDENING: Soft-delete aware unique index for roles.
CREATE UNIQUE INDEX idx_roles_unique_active_name ON roles (clinic_id, name) WHERE deleted_at IS NULL;

CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id SMALLINT NOT NULL REFERENCES permissions(id) ON DELETE RESTRICT,
    PRIMARY KEY (role_id, permission_id)
);

-- Join table for employees and roles
CREATE TABLE employee_roles (
    employee_profile_id UUID NOT NULL REFERENCES employees(profile_id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (employee_profile_id, role_id)
);
COMMENT ON TABLE employee_roles IS 'Assigns roles to employees.';

-- === HARDENING: Foundational Auditing System ===
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID,
    user_id UUID,
    action VARCHAR(50) NOT NULL,
    table_name TEXT NOT NULL,
    record_id UUID,
    old_record JSONB,
    new_record JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE audit_log IS 'Stores a detailed audit trail for all sensitive data changes.';
CREATE INDEX idx_audit_log_timestamp ON audit_log (timestamp);
CREATE INDEX idx_audit_log_clinic_id ON audit_log (clinic_id);
CREATE INDEX idx_audit_log_user_id ON audit_log (user_id);
CREATE INDEX idx_audit_log_table_record ON audit_log (table_name, record_id);

CREATE OR REPLACE FUNCTION log_change()
RETURNS TRIGGER AS $$
DECLARE
    audit_record audit_log;
    user_payload JSONB;
BEGIN
    BEGIN
        user_payload := current_setting('app.audit_context', true)::jsonb;
    EXCEPTION WHEN OTHERS THEN
        user_payload := '{}'::jsonb;
    END;
    audit_record = ROW(uuid_generate_v7(),(user_payload->>'clinic_id')::UUID,(user_payload->>'user_id')::UUID,TG_OP,TG_TABLE_NAME,NULL,NULL,NULL,NOW());
    IF (TG_OP = 'UPDATE') THEN
        audit_record.record_id := NEW.id;
        audit_record.old_record := to_jsonb(OLD);
        audit_record.new_record := to_jsonb(NEW);
    ELSIF (TG_OP = 'DELETE') THEN
        audit_record.record_id := OLD.id;
        audit_record.old_record := to_jsonb(OLD);
    ELSIF (TG_OP = 'INSERT') THEN
        audit_record.record_id := NEW.id;
        audit_record.new_record := to_jsonb(NEW);
    END IF;
    INSERT INTO audit_log VALUES (audit_record.*);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Apply audit trigger to the profiles table.
CREATE TRIGGER profiles_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON profiles
FOR EACH ROW EXECUTE FUNCTION log_change();

-- Apply the audit trigger to the 'employees' table to track staff changes.
CREATE TRIGGER employees_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON employees
FOR EACH ROW EXECUTE FUNCTION log_change();

-- Apply the audit trigger to the 'roles' table to track permission and role changes.
CREATE TRIGGER roles_audit_trigger
AFTER INSERT OR UPDATE OR DELETE ON roles
FOR EACH ROW EXECUTE FUNCTION log_change();

-- Apply timestamp triggers
CREATE TRIGGER set_timestamp BEFORE UPDATE ON profiles FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();

-- Seed permissions
INSERT INTO permissions (id, permission_key) VALUES
(1, 'employees.invite'), (2, 'employees.read'), (3, 'employees.update'), (4, 'employees.deactivate'),
(10, 'patients.create'), (11, 'patients.read'), (12, 'patients.update'), (13, 'patients.delete'),
(20, 'appointments.create'), (21, 'appointments.read'), (22, 'appointments.update'), (23, 'appointments.delete'),
(30, 'finance.invoice.create'), (31, 'finance.invoice.read'), (32, 'finance.payment.record'), (33, 'finance.reports.view'),
(40, 'roles.create'), (41, 'roles.read'), (42, 'roles.update'), (43, 'roles.delete');