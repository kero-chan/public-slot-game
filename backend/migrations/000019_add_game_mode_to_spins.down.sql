-- Remove game mode tracking columns
DROP INDEX IF EXISTS idx_spins_game_mode;
ALTER TABLE spins DROP COLUMN IF EXISTS game_mode_cost;
ALTER TABLE spins DROP COLUMN IF EXISTS game_mode;
