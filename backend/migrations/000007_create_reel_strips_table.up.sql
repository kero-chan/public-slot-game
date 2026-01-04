-- Create game_mode enum
CREATE TYPE game_mode AS ENUM (
    'base_game',
    'free_spins',
    'both'
);

-- Create reel_strips table for pre-generated reel strips
-- This optimizes performance by avoiding regeneration on every spin
CREATE TABLE reel_strips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Metadata
    game_mode game_mode NOT NULL, -- 'base_game' or 'free_spins' or 'both'
    reel_number INTEGER NOT NULL,   -- 0-4 (reel index for reels 1-5)
    -- Reel strip data (array of 1000 symbols for base_game, 100 for free_spins)
    strip_data JSONB NOT NULL,
    -- Verification and integrity
    checksum VARCHAR(64) NOT NULL UNIQUE,  -- SHA256 checksum for data integrity
    strip_length INTEGER NOT NULL,  -- Length of the strip (1000 or 100)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    -- Constraints
    CONSTRAINT reel_number_valid CHECK (reel_number >= 0 AND reel_number <= 4),
    CONSTRAINT strip_length_valid CHECK (strip_length > 0)
);
-- Indexes for fast querying
CREATE INDEX idx_reel_strips_active ON reel_strips(game_mode, reel_number, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_reel_strips_game_mode ON reel_strips(game_mode);
-- GIN index for JSONB queries (if needed for analytics)
CREATE INDEX idx_reel_strips_data ON reel_strips USING GIN(strip_data);
-- Add comment for documentation
COMMENT ON TABLE reel_strips IS 'Pre-generated reel strips for slot machine game. Each strip contains symbols in a specific order, optimizing spin performance by avoiding regeneration.';
COMMENT ON COLUMN reel_strips.game_mode IS 'Game mode: base_game (1000 symbols) or free_spins (100 symbols)';
COMMENT ON COLUMN reel_strips.reel_number IS 'Reel index (0-4) corresponding to reels 1-5';
COMMENT ON COLUMN reel_strips.strip_data IS 'JSON array of symbols in the reel strip';
COMMENT ON COLUMN reel_strips.checksum IS 'SHA256 checksum for verifying strip data integrity';
