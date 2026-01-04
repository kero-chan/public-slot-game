-- Revert: Remove object_name column (base_url is kept)

-- Drop the unique index on object_name
DROP INDEX IF EXISTS idx_assets_object_name;

-- Drop the object_name column
ALTER TABLE assets DROP COLUMN IF EXISTS object_name;

