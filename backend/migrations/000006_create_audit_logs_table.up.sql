-- Create audit_action enum
CREATE TYPE audit_action AS ENUM (
    'player_register',
    'player_login',
    'player_logout',
    'session_start',
    'session_end',
    'spin_execute',
    'free_spins_trigger',
    'balance_change',
    'settings_change'
);

-- Create audit_logs table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Actor
    player_id UUID REFERENCES players(id) ON DELETE SET NULL,
    -- Action
    action audit_action NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    -- Request data
    ip_address INET,
    user_agent TEXT,
    -- Details
    old_value JSONB,
    new_value JSONB,
    metadata JSONB,
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_audit_logs_player_id ON audit_logs(player_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
