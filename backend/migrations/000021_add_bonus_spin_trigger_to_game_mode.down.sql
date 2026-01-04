-- PostgreSQL does not support removing enum values directly
-- To rollback, you would need to:
-- 1. Create a new enum without the value
-- 2. Update all tables using the enum
-- 3. Drop the old enum and rename the new one
-- This is a destructive operation, so we leave a warning here

-- WARNING: This migration cannot be safely rolled back without data loss
-- If you need to rollback, ensure no data uses 'bonus_spin_trigger' value first

DO $$
BEGIN
    RAISE NOTICE 'WARNING: Removing enum values in PostgreSQL requires recreating the type. Ensure no data uses bonus_spin_trigger before proceeding.';
END $$;
