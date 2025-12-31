-- This migration safely tears down the unified IAM and Profiles schema in the reverse order of creation.

-- Drop audit components first
DROP TRIGGER IF EXISTS profiles_audit_trigger ON profiles;
DROP FUNCTION IF EXISTS log_change();
DROP TABLE IF EXISTS audit_log;

-- Drop triggers first as they depend on the tables and functions.
DROP TRIGGER IF EXISTS set_timestamp ON roles;
DROP TRIGGER IF EXISTS set_timestamp ON employees;
DROP TRIGGER IF EXISTS set_timestamp ON profiles;

-- Drop join tables before primary tables.
DROP TABLE IF EXISTS employee_roles;
DROP TABLE IF EXISTS role_permissions;

-- Drop primary RBAC tables.
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;

-- Drop the employees table before the profiles table due to the foreign key relationship.
DROP TABLE IF EXISTS employees;

-- Drop the unified profiles table.
DROP TABLE IF EXISTS profiles;

-- Drop the custom ENUM types.
DROP TYPE IF EXISTS employee_status;
DROP TYPE IF EXISTS profile_status;