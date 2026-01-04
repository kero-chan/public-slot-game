-- Remove Dual Commitment Protocol fields from pf_sessions table
ALTER TABLE pf_sessions
DROP COLUMN IF EXISTS theta_commitment,
DROP COLUMN IF EXISTS theta_seed,
DROP COLUMN IF EXISTS theta_verified;
