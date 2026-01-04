-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_prevent_session_audit_delete ON session_audits;
DROP TRIGGER IF EXISTS trigger_prevent_session_audit_update ON session_audits;
DROP TRIGGER IF EXISTS trigger_prevent_spin_log_delete ON spin_logs;
DROP TRIGGER IF EXISTS trigger_prevent_spin_log_update ON spin_logs;

-- Drop function
DROP FUNCTION IF EXISTS prevent_spin_log_modification();

-- Drop tables in reverse order of creation (due to foreign key constraints)
DROP TABLE IF EXISTS session_audits;
DROP TABLE IF EXISTS spin_logs;
DROP TABLE IF EXISTS pf_sessions;
