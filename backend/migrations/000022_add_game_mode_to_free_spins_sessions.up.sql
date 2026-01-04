-- Add reel_strip_config_id column to free_spins_sessions table
ALTER TABLE free_spins_sessions ADD COLUMN reel_strip_config_id UUID DEFAULT NULL;

-- Add foreign key constraint (optional - config might not exist)
ALTER TABLE free_spins_sessions ADD CONSTRAINT fk_free_spins_reel_strip_config
    FOREIGN KEY (reel_strip_config_id) REFERENCES reel_strip_configs(id) ON DELETE SET NULL;

-- Add comment
COMMENT ON COLUMN free_spins_sessions.reel_strip_config_id IS 'Reference to reel strip config used for this free spins session (null = use default config)';
