Alter table free_spins_sessions add column if not exists session_id UUID REFERENCES game_sessions(id) ON DELETE CASCADE;
