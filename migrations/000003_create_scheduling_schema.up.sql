-- This migration establishes the tables required for the core scheduling engine,
-- now correctly referencing the unified 'profiles' and 'employees' tables.

-- Add scheduling-specific configuration to the clinics table.
ALTER TABLE clinics
    ADD COLUMN IF NOT EXISTS slot_duration INTERVAL NOT NULL DEFAULT '30 minutes';

COMMENT ON COLUMN clinics.slot_duration IS 'The fundamental time unit for the clinic''s calendar grid, e.g., 15 minutes, 30 minutes.';

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(19, 4) NOT NULL DEFAULT 0.00,
    slot_multiple INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_services_clinic_name UNIQUE (clinic_id, name),
    CONSTRAINT chk_slot_multiple_positive CHECK (slot_multiple > 0)
);
COMMENT ON TABLE services IS 'Defines the clinical procedures offered. Duration is based on a multiple of the clinic''s slot_duration.';

CREATE TABLE doctor_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES employees(profile_id) ON DELETE CASCADE,
    day_of_week INT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_day_of_week_range CHECK (day_of_week >= 0 AND day_of_week <= 6),
    CONSTRAINT chk_schedule_times CHECK (end_time > start_time),
    EXCLUDE USING GIST (
        doctor_id WITH =,
        day_of_week WITH =,
        timerange(start_time, end_time, '[]') WITH &&
    )
);
COMMENT ON TABLE doctor_schedules IS 'Stores recurring weekly working hours for doctors (employees).';

CREATE TABLE appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    patient_id UUID NOT NULL REFERENCES profiles(id) ON DELETE RESTRICT,
    doctor_id UUID NOT NULL REFERENCES employees(profile_id) ON DELETE RESTRICT,
    service_id UUID REFERENCES services(id) ON DELETE SET NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    guest_management_token VARCHAR(255) UNIQUE,
    guest_management_token_expires_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'SCHEDULED',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_appointment_times CHECK (end_time > start_time),
    EXCLUDE USING GIST (
        doctor_id WITH =,
        tstzrange(start_time, end_time) WITH &&
    )
);
COMMENT ON TABLE appointments IS 'Stores all scheduled appointments. Links employees (doctors) to profiles (patients).';

-- Apply timestamp triggers
CREATE TRIGGER set_timestamp BEFORE UPDATE ON services FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp BEFORE UPDATE ON doctor_schedules FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
CREATE TRIGGER set_timestamp BEFORE UPDATE ON appointments FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();