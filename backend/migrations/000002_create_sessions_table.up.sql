-- Create game_sessions table
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Session data
    bet_amount DECIMAL(10, 2) NOT NULL,
    starting_balance DECIMAL(15, 2) NOT NULL,
    ending_balance DECIMAL(15, 2),
    -- Statistics
    total_spins INTEGER DEFAULT 0,
    total_wagered DECIMAL(15, 2) DEFAULT 0.00,
    total_won DECIMAL(15, 2) DEFAULT 0.00,
    net_change DECIMAL(15, 2) DEFAULT 0.00,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    -- Constraints
    CONSTRAINT bet_amount_positive CHECK (bet_amount > 0)
);
-- Create indexes
CREATE INDEX idx_sessions_player_id ON game_sessions(player_id);
CREATE INDEX idx_sessions_created_at ON game_sessions(created_at);
CREATE INDEX idx_sessions_ended_at ON game_sessions(ended_at);
CREATE INDEX idx_incomplete_sessions ON game_sessions(player_id) WHERE ended_at IS NULL;
