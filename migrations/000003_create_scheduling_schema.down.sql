-- This migration safely tears down the scheduling schema in reverse order of creation.

-- Drop triggers first.
DROP TRIGGER IF EXISTS set_timestamp ON appointments;
DROP TRIGGER IF EXISTS set_timestamp ON doctor_schedules;
DROP TRIGGER IF EXISTS set_timestamp ON services;

-- Drop tables that have foreign keys pointing to other tables.
DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS doctor_schedules;
DROP TABLE IF EXISTS services;

-- Revert changes to the clinics table by dropping the added column.
ALTER TABLE clinics
    DROP COLUMN IF EXISTS slot_duration;