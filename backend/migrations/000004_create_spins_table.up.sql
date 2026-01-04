-- Create spins table
CREATE TABLE spins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES game_sessions(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Spin data
    bet_amount DECIMAL(10, 2) NOT NULL,
    balance_before DECIMAL(15, 2) NOT NULL,
    balance_after DECIMAL(15, 2) NOT NULL,
    -- Grid state (JSON)
    grid JSONB NOT NULL,
    cascades JSONB,
    -- Win data
    total_win DECIMAL(15, 2) DEFAULT 0.00,
    scatter_count INTEGER DEFAULT 0,
    -- Free spins
    is_free_spin BOOLEAN DEFAULT FALSE,
    free_spins_session_id UUID REFERENCES free_spins_sessions(id),
    free_spins_triggered BOOLEAN DEFAULT FALSE,
    reel_positions JSONB,
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Constraints
    CONSTRAINT scatter_count_valid CHECK (scatter_count >= 0 AND scatter_count <= 20)
);
-- Create indexes
CREATE INDEX idx_spins_session_id ON spins(session_id);
CREATE INDEX idx_spins_player_id ON spins(player_id);
CREATE INDEX idx_spins_created_at ON spins(created_at);
CREATE INDEX idx_spins_is_free_spin ON spins(is_free_spin);
CREATE INDEX idx_spins_free_spins_session_id ON spins(free_spins_session_id);
CREATE INDEX idx_spins_player_created ON spins(player_id, created_at DESC);
CREATE INDEX idx_spins_session_created ON spins(session_id, created_at);
-- GIN index for JSONB grid queries
CREATE INDEX idx_spins_grid ON spins USING GIN(grid);
CREATE INDEX idx_spins_cascades ON spins USING GIN(cascades);
-- Create trigger to update player statistics on new spin
-- CREATE OR REPLACE FUNCTION update_player_statistics_trigger()
-- RETURNS TRIGGER AS $$
-- BEGIN
--     UPDATE players
--     SET
--         total_spins = total_spins + 1,
--         total_wagered = total_wagered + NEW.bet_amount,
--         total_won = total_won + NEW.total_win,
--         updated_at = CURRENT_TIMESTAMP
--     WHERE id = NEW.player_id;
--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;
-- CREATE TRIGGER trigger_update_player_statistics
-- AFTER INSERT ON spins
-- FOR EACH ROW
-- EXECUTE FUNCTION update_player_statistics_trigger();
