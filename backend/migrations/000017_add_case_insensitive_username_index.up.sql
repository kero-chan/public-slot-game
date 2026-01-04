-- Add case-insensitive unique indexes for username and email
-- This prevents race conditions where two concurrent requests could create
-- accounts with the same username in different cases (e.g., "Test" and "test")

-- Drop existing indexes
DROP INDEX IF EXISTS idx_players_username_game_id;
DROP INDEX IF EXISTS idx_players_username_null_game;
DROP INDEX IF EXISTS idx_players_email_game_id;
DROP INDEX IF EXISTS idx_players_email_null_game;

-- Create case-insensitive unique index for username + game_id (when game_id IS NOT NULL)
-- Ensures: same username (case-insensitive) cannot exist twice for the same game
CREATE UNIQUE INDEX idx_players_username_game_id
    ON players (LOWER(username), game_id)
    WHERE game_id IS NOT NULL;

-- Create case-insensitive unique index for username (when game_id IS NULL)
-- Ensures: only ONE cross-game account per username (case-insensitive)
CREATE UNIQUE INDEX idx_players_username_null_game
    ON players (LOWER(username))
    WHERE game_id IS NULL;

-- Create case-insensitive unique index for email + game_id (when game_id IS NOT NULL)
-- Ensures: same email (case-insensitive) cannot exist twice for the same game
CREATE UNIQUE INDEX idx_players_email_game_id
    ON players (LOWER(email), game_id)
    WHERE game_id IS NOT NULL;

-- Create case-insensitive unique index for email (when game_id IS NULL)
-- Ensures: only ONE cross-game account per email (case-insensitive)
CREATE UNIQUE INDEX idx_players_email_null_game
    ON players (LOWER(email))
    WHERE game_id IS NULL;
