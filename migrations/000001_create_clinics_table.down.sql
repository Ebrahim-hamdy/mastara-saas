-- This migration safely tears down the 'clinics' table and its related components.

-- Drop the trigger first, as it depends on the table.
DROP TRIGGER IF EXISTS set_timestamp ON clinics;

-- Drop the function that the trigger uses.
DROP FUNCTION IF EXISTS trigger_set_timestamp();

-- Drop the table itself.
DROP TABLE IF EXISTS clinics;

-- Drop the extension. This is safe as long as no other tables are using it.
DROP EXTENSION IF EXISTS "pg_uuidv7";