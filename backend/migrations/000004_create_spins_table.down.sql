-- Drop spins table
DROP TRIGGER IF EXISTS trigger_update_player_statistics ON spins;
-- DROP FUNCTION IF EXISTS update_player_statistics_trigger();
DROP TABLE IF EXISTS spins CASCADE;
