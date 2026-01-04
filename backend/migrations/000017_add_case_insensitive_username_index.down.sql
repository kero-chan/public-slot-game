-- Rollback: restore original case-sensitive indexes

-- Drop case-insensitive indexes
DROP INDEX IF EXISTS idx_players_username_game_id;
DROP INDEX IF EXISTS idx_players_username_null_game;
DROP INDEX IF EXISTS idx_players_email_game_id;
DROP INDEX IF EXISTS idx_players_email_null_game;

-- Restore original case-sensitive indexes
CREATE UNIQUE INDEX idx_players_username_game_id
    ON players (username, game_id)
    WHERE game_id IS NOT NULL;

CREATE UNIQUE INDEX idx_players_username_null_game
    ON players (username)
    WHERE game_id IS NULL;

CREATE UNIQUE INDEX idx_players_email_game_id
    ON players (email, game_id)
    WHERE game_id IS NOT NULL;

CREATE UNIQUE INDEX idx_players_email_null_game
    ON players (email)
    WHERE game_id IS NULL;
