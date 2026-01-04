-- Add game_id column to players table (nullable foreign key to games)
ALTER TABLE players
ADD COLUMN game_id UUID REFERENCES games(id) ON DELETE SET NULL;

-- Drop existing unique indexes (GORM created these)
DROP INDEX IF EXISTS idx_players_username;
DROP INDEX IF EXISTS idx_players_email;

-- Remove old unique constraints (try all possible naming conventions)
ALTER TABLE players DROP CONSTRAINT IF EXISTS uni_players_username;
ALTER TABLE players DROP CONSTRAINT IF EXISTS uni_players_email;
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_username_key;
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_email_key;

-- Create partial unique index for username + game_id (when game_id IS NOT NULL)
-- Ensures: same username cannot exist twice for the same game
CREATE UNIQUE INDEX idx_players_username_game_id
    ON players (username, game_id)
    WHERE game_id IS NOT NULL;

-- Create partial unique index for username (when game_id IS NULL)
-- Ensures: only ONE cross-game account per username
CREATE UNIQUE INDEX idx_players_username_null_game
    ON players (username)
    WHERE game_id IS NULL;

-- Create partial unique index for email + game_id (when game_id IS NOT NULL)
-- Ensures: same email cannot exist twice for the same game
CREATE UNIQUE INDEX idx_players_email_game_id
    ON players (email, game_id)
    WHERE game_id IS NOT NULL;

-- Create partial unique index for email (when game_id IS NULL)
-- Ensures: only ONE cross-game account per email
CREATE UNIQUE INDEX idx_players_email_null_game
    ON players (email)
    WHERE game_id IS NULL;

-- Create index for game_id lookups
CREATE INDEX idx_players_game_id ON players (game_id);
