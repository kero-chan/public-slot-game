ALTER TABLE public.reel_strip_configs ADD COLUMN IF NOT EXISTS options jsonb DEFAULT '{}'::jsonb NULL;
