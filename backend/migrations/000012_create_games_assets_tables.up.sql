-- Create games table
CREATE TABLE games (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- Create assets table
CREATE TABLE assets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(100) NOT NULL UNIQUE,
    description         TEXT,
    object_name         VARCHAR(100) NOT NULL,
    base_url            TEXT NOT NULL,
    spritesheet_json    JSONB NOT NULL,
    images              JSONB NOT NULL,
    is_active           BOOLEAN DEFAULT true,
    created_at          TIMESTAMP DEFAULT NOW(),
    updated_at          TIMESTAMP DEFAULT NOW()
);

-- Create game_configs table
CREATE TABLE game_configs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id         UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    asset_id        UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_games_is_active ON games (is_active);
CREATE INDEX idx_assets_is_active ON assets (is_active);
CREATE INDEX idx_game_configs_game_id ON game_configs (game_id);
CREATE INDEX idx_game_configs_asset_id ON game_configs (asset_id);

-- Partial unique index: only one active config per game
CREATE UNIQUE INDEX idx_game_configs_active_game
    ON game_configs (game_id) WHERE is_active = true;
