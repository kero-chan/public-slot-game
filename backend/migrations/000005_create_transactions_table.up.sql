-- Create transaction_type enum
CREATE TYPE transaction_type AS ENUM (
    'deposit',
    'withdrawal',
    'bet',
    'win',
    'refund',
    'bonus',
    'adjustment'
);

-- Create transactions table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    -- Transaction data
    type transaction_type NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    balance_before DECIMAL(15, 2) NOT NULL,
    balance_after DECIMAL(15, 2) NOT NULL,
    -- Related records
    spin_id UUID REFERENCES spins(id) ON DELETE SET NULL,
    session_id UUID REFERENCES game_sessions(id) ON DELETE SET NULL,
    -- Metadata
    description TEXT,
    metadata JSONB,
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Constraints
    CONSTRAINT balance_after_correct CHECK (
        balance_after = balance_before + CASE
            WHEN type IN ('deposit', 'win', 'bonus', 'refund') THEN amount
            WHEN type IN ('withdrawal', 'bet') THEN -amount
            ELSE 0
        END
    )
);
-- Create indexes
CREATE INDEX idx_transactions_player_id ON transactions(player_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_spin_id ON transactions(spin_id);
CREATE INDEX idx_transactions_session_id ON transactions(session_id);
CREATE INDEX idx_transactions_player_type_created ON transactions(player_id, type, created_at DESC);
