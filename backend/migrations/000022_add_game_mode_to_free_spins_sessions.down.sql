-- Remove foreign key constraint
ALTER TABLE free_spins_sessions DROP CONSTRAINT IF EXISTS fk_free_spins_reel_strip_config;

-- Remove reel_strip_config_id column from free_spins_sessions table
ALTER TABLE free_spins_sessions DROP COLUMN IF EXISTS reel_strip_config_id;
