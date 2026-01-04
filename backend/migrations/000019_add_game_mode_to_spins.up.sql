-- Add game_mode column to track special spin modes
-- Values: NULL (normal spin), 'free_spin_trigger', 'wild_spin_trigger', 'bonus_spin_trigger'
ALTER TABLE spins ADD COLUMN game_mode VARCHAR(32) DEFAULT NULL;

-- Add game_mode_cost column to track the cost paid for the mode
ALTER TABLE spins ADD COLUMN game_mode_cost DECIMAL(10, 2) DEFAULT NULL;

-- Add partial index for querying by game mode (only indexes non-null values)
CREATE INDEX idx_spins_game_mode ON spins(game_mode) WHERE game_mode IS NOT NULL;
