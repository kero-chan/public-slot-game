-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_update_reel_strip_config_updated_at ON reel_strip_configs;
DROP FUNCTION IF EXISTS update_reel_strip_config_updated_at();

-- Drop player_reel_strip_assignments table and its indexes
DROP INDEX IF EXISTS idx_player_assignments_player;
DROP INDEX IF EXISTS idx_player_assignments_active;
DROP INDEX IF EXISTS idx_player_active_assignment;
DROP TABLE IF EXISTS player_reel_strip_assignments;

-- Drop reel_strip_configs table and its indexes
DROP INDEX IF EXISTS idx_reel_strip_configs_game_mode;
DROP INDEX IF EXISTS idx_reel_strip_configs_default;
DROP INDEX IF EXISTS idx_reel_strip_configs_active;
DROP TABLE IF EXISTS reel_strip_configs;
