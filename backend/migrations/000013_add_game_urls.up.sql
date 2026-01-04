-- Add dev_url and prod_url columns to games table
ALTER TABLE games ADD COLUMN dev_url TEXT;
ALTER TABLE games ADD COLUMN prod_url TEXT;
