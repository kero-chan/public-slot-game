-- Create player_sessions table for tracking active login sessions
-- Supports single-device login per player per game with force logout capability

CREATE TABLE IF NOT EXISTS player_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    game_id UUID REFERENCES games(id) ON DELETE CASCADE,

    -- Session token (unique identifier for this session)
    session_token VARCHAR(64) NOT NULL UNIQUE,

    -- Device/client info
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Session status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    logged_out_at TIMESTAMP WITH TIME ZONE,

    -- Logout reason (null = still active, 'manual' = user logout, 'forced' = new device login, 'expired' = token expired)
    logout_reason VARCHAR(20)
);

-- Index for fast lookup by player + game (for single-device enforcement)
CREATE UNIQUE INDEX idx_player_sessions_active_player_game
    ON player_sessions (player_id, game_id)
    WHERE is_active = true AND game_id IS NOT NULL;

-- Index for cross-game players (game_id IS NULL)
CREATE UNIQUE INDEX idx_player_sessions_active_player_null_game
    ON player_sessions (player_id)
    WHERE is_active = true AND game_id IS NULL;

-- Index for session token lookup
CREATE INDEX idx_player_sessions_token ON player_sessions (session_token) WHERE is_active = true;

-- Index for cleanup of expired sessions
CREATE INDEX idx_player_sessions_expires_at ON player_sessions (expires_at) WHERE is_active = true;

-- Index for player session history
CREATE INDEX idx_player_sessions_player_id ON player_sessions (player_id);

COMMENT ON TABLE player_sessions IS 'Tracks active player login sessions for single-device enforcement';
COMMENT ON COLUMN player_sessions.session_token IS 'Unique token stored in JWT and Redis for session validation';
COMMENT ON COLUMN player_sessions.logout_reason IS 'Reason for logout: manual, forced (new device), expired';
