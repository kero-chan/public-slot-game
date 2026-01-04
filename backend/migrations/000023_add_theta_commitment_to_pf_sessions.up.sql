-- Add Dual Commitment Protocol fields to pf_sessions table
-- theta_commitment: SHA256(theta_seed) - client's commitment sent BEFORE seeing server_seed
-- theta_seed: Client's session seed - revealed on first spin
-- theta_verified: True after theta_seed is verified on first spin

ALTER TABLE pf_sessions
ADD COLUMN theta_commitment VARCHAR(64),
ADD COLUMN theta_seed VARCHAR(64),
ADD COLUMN theta_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Add comment to document the Dual Commitment Protocol
COMMENT ON COLUMN pf_sessions.theta_commitment IS 'SHA256(theta_seed) - client sends this BEFORE server generates server_seed';
COMMENT ON COLUMN pf_sessions.theta_seed IS 'Client session seed - revealed on first spin and verified against theta_commitment';
COMMENT ON COLUMN pf_sessions.theta_verified IS 'True after SHA256(theta_seed) === theta_commitment verification passes';
