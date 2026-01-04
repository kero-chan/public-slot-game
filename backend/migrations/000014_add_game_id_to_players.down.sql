-- Drop new indexes
DROP INDEX IF EXISTS idx_players_username_game_id;
DROP INDEX IF EXISTS idx_players_username_null_game;
DROP INDEX IF EXISTS idx_players_email_game_id;
DROP INDEX IF EXISTS idx_players_email_null_game;
DROP INDEX IF EXISTS idx_players_game_id;

-- Remove game_id column
ALTER TABLE players DROP COLUMN IF EXISTS game_id;

-- Restore original unique indexes
CREATE UNIQUE INDEX idx_players_username ON players (username);
CREATE UNIQUE INDEX idx_players_email ON players (email);
