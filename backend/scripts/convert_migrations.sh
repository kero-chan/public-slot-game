#!/bin/bash
set -e

# Script to convert old migrations to golang-migrate format
# Old format: 0001_name.sql
# New format: 000001_name.up.sql and 000001_name.down.sql

OLD_MIGRATIONS_DIR="migrations"
NEW_MIGRATIONS_DIR="migrations_new"

echo "🔄 Converting migrations to golang-migrate format..."

# Create new migrations directory
mkdir -p "$NEW_MIGRATIONS_DIR"

# Counter for sequential numbering
counter=1

# Process each old migration file
for old_file in "$OLD_MIGRATIONS_DIR"/*.sql; do
    if [ ! -f "$old_file" ]; then
        continue
    fi

    # Extract filename without path
    filename=$(basename "$old_file")

    # Extract the name part (remove 000X_ prefix and .sql suffix)
    name=$(echo "$filename" | sed -E 's/^[0-9]+_(.*)\.sql$/\1/')

    # Create new filenames with 6-digit padding
    new_number=$(printf "%06d" $counter)
    up_file="${NEW_MIGRATIONS_DIR}/${new_number}_${name}.up.sql"
    down_file="${NEW_MIGRATIONS_DIR}/${new_number}_${name}.down.sql"

    echo "  Converting: $filename -> ${new_number}_${name}.{up,down}.sql"

    # Copy content to UP migration
    cp "$old_file" "$up_file"

    # Create DOWN migration based on the table name
    case "$name" in
        "create_players_table")
            cat > "$down_file" << 'EOF'
-- Drop players table
DROP TRIGGER IF EXISTS trigger_update_players_timestamp ON players;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS players CASCADE;
EOF
            ;;
        "create_sessions_table")
            cat > "$down_file" << 'EOF'
-- Drop game_sessions table
DROP TABLE IF EXISTS game_sessions CASCADE;
EOF
            ;;
        "create_free_spins_sessions_table")
            cat > "$down_file" << 'EOF'
-- Drop free_spins_sessions table
DROP TABLE IF EXISTS free_spins_sessions CASCADE;
EOF
            ;;
        "create_spins_table")
            cat > "$down_file" << 'EOF'
-- Drop spins table
DROP TRIGGER IF EXISTS trigger_update_player_statistics ON spins;
DROP FUNCTION IF EXISTS update_player_statistics_trigger();
DROP TABLE IF EXISTS spins CASCADE;
EOF
            ;;
        "create_transactions_table")
            cat > "$down_file" << 'EOF'
-- Drop transactions table
DROP TABLE IF EXISTS transactions CASCADE;
EOF
            ;;
        "create_audit_logs_table")
            cat > "$down_file" << 'EOF'
-- Drop audit_logs table
DROP TABLE IF EXISTS audit_logs CASCADE;
EOF
            ;;
        "create_reel_strips_table")
            cat > "$down_file" << 'EOF'
-- Drop reel_strips table
DROP TABLE IF EXISTS reel_strips CASCADE;
-- Note: Extensions are not dropped as they might be used by other tables
EOF
            ;;
        *)
            # Generic down migration
            echo "-- Rollback for $name" > "$down_file"
            echo "-- TODO: Add rollback SQL" >> "$down_file"
            ;;
    esac

    counter=$((counter + 1))
done

echo ""
echo "✅ Conversion complete!"
echo "   Old migrations: $OLD_MIGRATIONS_DIR/"
echo "   New migrations: $NEW_MIGRATIONS_DIR/"
echo ""
echo "Next steps:"
echo "1. Review the DOWN migrations in $NEW_MIGRATIONS_DIR/"
echo "2. Backup old migrations: mv migrations migrations_backup"
echo "3. Replace with new: mv migrations_new migrations"
echo "4. Test: make migrate"
