#!/bin/bash
set -e

# Script to replace uuid-ossp extension with built-in gen_random_uuid()
# PostgreSQL 13+ has gen_random_uuid() built-in, no extension needed

echo "🔧 Fixing UUID extensions in migrations..."

MIGRATIONS_DIR="migrations"

# Find all .up.sql files
for file in "$MIGRATIONS_DIR"/*.up.sql; do
    if [ -f "$file" ]; then
        echo "  Processing: $(basename $file)"

        # Replace uuid-ossp extension with pgcrypto (for gen_random_uuid)
        sed -i.bak 's/CREATE EXTENSION IF NOT EXISTS "uuid-ossp";/-- UUID generation uses gen_random_uuid() (PostgreSQL 13+)\nCREATE EXTENSION IF NOT EXISTS "pgcrypto";/g' "$file"

        # Replace uuid_generate_v4() with gen_random_uuid()
        sed -i.bak 's/uuid_generate_v4()/gen_random_uuid()/g' "$file"

        # Remove backup files
        rm -f "${file}.bak"
    fi
done

echo ""
echo "✅ UUID extensions fixed!"
echo ""
echo "Changes made:"
echo "  - Removed 'uuid-ossp' extension dependency"
echo "  - Replaced uuid_generate_v4() with gen_random_uuid()"
echo "  - Added pgcrypto extension (for gen_random_uuid and password hashing)"
echo ""
echo "Now run: make migrate"
