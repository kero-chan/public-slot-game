-- Remove dev_url and prod_url columns from games table
ALTER TABLE games DROP COLUMN IF EXISTS dev_url;
ALTER TABLE games DROP COLUMN IF EXISTS prod_url;
