-- Create free_spins_sessions table
CREATE TABLE free_spins_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    triggered_by_spin_id UUID,

    -- Free spins data
    scatter_count INTEGER NOT NULL,
    total_spins_awarded INTEGER NOT NULL,
    spins_completed INTEGER DEFAULT 0,
    remaining_spins INTEGER NOT NULL,
    locked_bet_amount DECIMAL(10, 2) NOT NULL,
    -- Win tracking
    total_won DECIMAL(15, 2) DEFAULT 0.00,
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    is_completed BOOLEAN DEFAULT FALSE,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    lock_version INTEGER DEFAULT 0 NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    -- Constraints
    CONSTRAINT scatter_count_min CHECK (scatter_count >= 3),
    CONSTRAINT remaining_spins_non_negative CHECK (remaining_spins >= 0)
);
-- Create indexes
CREATE INDEX idx_free_spins_sessions_player_id ON free_spins_sessions(player_id);
CREATE INDEX idx_free_spins_sessions_is_active ON free_spins_sessions(is_active);
CREATE INDEX idx_free_spins_sessions_created_at ON free_spins_sessions(created_at);
CREATE INDEX idx_active_free_spins_sessions ON free_spins_sessions(player_id) WHERE is_active = TRUE;
