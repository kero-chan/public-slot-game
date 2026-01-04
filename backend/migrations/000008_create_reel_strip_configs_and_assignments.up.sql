-- Create reel_strip_configs table
CREATE TABLE IF NOT EXISTS reel_strip_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    game_mode game_mode NOT NULL,
    description TEXT,

    -- Reel strip references
    reel0_strip_id UUID NOT NULL REFERENCES reel_strips(id),
    reel1_strip_id UUID NOT NULL REFERENCES reel_strips(id),
    reel2_strip_id UUID NOT NULL REFERENCES reel_strips(id),
    reel3_strip_id UUID NOT NULL REFERENCES reel_strips(id),
    reel4_strip_id UUID NOT NULL REFERENCES reel_strips(id),

    -- Metadata
    target_rtp DECIMAL(6,2),
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    activated_at TIMESTAMP WITH TIME ZONE,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    notes TEXT,
    options JSONB DEFAULT '{}'::jsonb
);

-- Create indexes for reel_strip_configs
CREATE INDEX idx_reel_strip_configs_active ON reel_strip_configs(is_active, game_mode);
CREATE INDEX idx_reel_strip_configs_default ON reel_strip_configs(is_default, game_mode) WHERE is_default = true;
CREATE INDEX idx_reel_strip_configs_game_mode ON reel_strip_configs(game_mode);

-- Create player_reel_strip_assignments table
CREATE TABLE IF NOT EXISTS player_reel_strip_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,

    -- Config assignments
    base_game_config_id UUID REFERENCES reel_strip_configs(id),
    free_spins_config_id UUID REFERENCES reel_strip_configs(id),

    -- Assignment metadata
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_by VARCHAR(100),
    reason VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

-- Create indexes for player_reel_strip_assignments
CREATE UNIQUE INDEX idx_player_active_assignment ON player_reel_strip_assignments(player_id) WHERE is_active = true;
CREATE INDEX idx_player_assignments_active ON player_reel_strip_assignments(is_active);
CREATE INDEX idx_player_assignments_player ON player_reel_strip_assignments(player_id);

-- Add comments
COMMENT ON TABLE reel_strip_configs IS 'Manages reel strip configurations for different game versions and player segments';
COMMENT ON TABLE player_reel_strip_assignments IS 'Assigns specific reel strip configurations to players for A/B testing or VIP experiences';
COMMENT ON COLUMN reel_strip_configs.is_default IS 'Default configuration for new players (only one per game_mode should be true)';
COMMENT ON COLUMN player_reel_strip_assignments.reason IS 'Reason for assignment (e.g., "A/B Test Group A", "VIP Player", "High Roller")';
