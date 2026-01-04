-- Drop players table
DROP TRIGGER IF EXISTS trigger_update_players_timestamp ON players;
-- DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS players CASCADE;
