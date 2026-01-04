-- Add object_name column to assets table
-- object_name is the folder name in storage (e.g., "bamboo-abc123")
-- base_url is kept and set once at creation time (not dynamically computed)

-- First add the object_name column as nullable
ALTER TABLE assets ADD COLUMN IF NOT EXISTS object_name VARCHAR(100);

-- For existing records: extract object_name from name (lowercase, replace spaces with dashes)
-- Add row number suffix to handle potential duplicates
WITH numbered_assets AS (
    SELECT id, name,
           lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g')) as base_name,
           ROW_NUMBER() OVER (PARTITION BY lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g')) ORDER BY created_at) as rn
    FROM assets
    WHERE object_name IS NULL
)
UPDATE assets
SET object_name = CASE
    WHEN numbered_assets.rn = 1 THEN numbered_assets.base_name
    ELSE numbered_assets.base_name || '-v' || numbered_assets.rn
END
FROM numbered_assets
WHERE assets.id = numbered_assets.id;

-- Now make the column not null and add unique index
ALTER TABLE assets ALTER COLUMN object_name SET NOT NULL;

-- Add unique index
CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_object_name ON assets(object_name);
