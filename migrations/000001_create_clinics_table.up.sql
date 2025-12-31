-- This migration sets up the foundational extensions and the core 'clinics' table,
-- which is the root of our multi-tenancy architecture.

-- Enable the pg_uuidv7 extension to generate time-ordered, sortable UUIDs (our standard for all PKs).
CREATE EXTENSION IF NOT EXISTS "pg_uuidv7";

-- Enable btree_gist to allow GIST indexes to work with standard data types
-- in EXCLUDE constraints (e.g., for doctor_id, day_of_week). This is critical.
CREATE EXTENSION IF NOT EXISTS "btree_gist";


CREATE TABLE clinics (
    -- Core Identity
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,

    -- Contact Information
    phone_number VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,

    -- Physical Address (Structured for querying and formatting)
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    city VARCHAR(100) NOT NULL,
    state_province VARCHAR(100),
    postal_code VARCHAR(20) NOT NULL,
    country_code CHAR(2) NOT NULL, -- ISO 3166-1 alpha-2 country code (e.g., 'EG', 'SA')

    -- Operational Settings
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC', -- IANA Time Zone (e.g., 'Africa/Cairo')
    currency CHAR(3) NOT NULL DEFAULT 'EGP',      -- ISO 4217 currency code

    -- SaaS Subscription Information
    -- This will link to a future 'subscription_plans' table.
    subscription_plan_id UUID,
    subscription_status VARCHAR(50) NOT NULL DEFAULT 'trial', -- e.g., 'trial', 'active', 'suspended', 'cancelled'

    -- Flexible settings storage for future needs without schema changes.
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Standard Audit Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Add comments to explain the purpose of columns and the table itself.
COMMENT ON TABLE clinics IS 'Represents a single tenant (a dental clinic) in the SaaS platform.';
COMMENT ON COLUMN clinics.timezone IS 'The IANA timezone name for the clinic, used for displaying local times.';
COMMENT ON COLUMN clinics.settings IS 'Flexible JSONB field for clinic-specific settings like logo URL, branding colors, etc.';

-- === HARDENING: Soft-delete aware unique index ===
CREATE UNIQUE INDEX idx_clinics_unique_active_email ON clinics (email) WHERE deleted_at IS NULL;

-- Create indexes for performance on frequently queried columns.
CREATE INDEX idx_clinics_name ON clinics(name);
CREATE INDEX idx_clinics_subscription_status ON clinics(subscription_status);

-- A trigger to automatically update the 'updated_at' timestamp on any change.
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON clinics
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();