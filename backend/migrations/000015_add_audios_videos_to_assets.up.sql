-- Add audios and videos columns to assets table
ALTER TABLE assets ADD COLUMN audios JSONB DEFAULT '{}';
ALTER TABLE assets ADD COLUMN videos JSONB DEFAULT '{}';

-- Add comment for documentation
COMMENT ON COLUMN assets.audios IS 'JSON mapping of audio names to file paths (e.g., {"background_music": "audios/background_music.mp3"})';
COMMENT ON COLUMN assets.videos IS 'JSON mapping of video names to file paths (e.g., {"jackpot_intro": "videos/jackpot_intro.mp4"})';
