-- This migration safely tears down the entire IAM schema in the reverse order of creation
-- to respect foreign key constraints.

-- Drop triggers first
DROP TRIGGER IF EXISTS set_timestamp ON users;
DROP TRIGGER IF EXISTS set_timestamp ON roles;

-- Drop join tables
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;

-- Drop primary tables
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;

-- Drop the custom ENUM type
DROP TYPE IF EXISTS user_status;