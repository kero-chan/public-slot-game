-- Remove audios and videos columns from assets table
ALTER TABLE assets DROP COLUMN IF EXISTS audios;
ALTER TABLE assets DROP COLUMN IF EXISTS videos;
