#!/bin/bash
set -e

# Get the script directory
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to project root
cd "$PROJECT_ROOT"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default database connection if not set
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-slotmachine}"

# PostgreSQL connection string
export PGPASSWORD="$DB_PASSWORD"
PSQL_CMD="psql -h $DB_HOST -p $DB_PORT -U $DB_USER"

echo "🗄️  Running database migrations..."
echo "   Database: $DB_NAME"
echo "   Host: $DB_HOST:$DB_PORT"
echo ""

# Create database if it doesn't exist
$PSQL_CMD -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
$PSQL_CMD -c "CREATE DATABASE $DB_NAME"

# Run migrations in order
for migration in migrations/*.sql; do
    if [ -f "$migration" ]; then
        echo "→ Running $(basename $migration)..."
        $PSQL_CMD -d $DB_NAME -f "$migration"
    fi
done

echo ""
echo "✓ All migrations completed successfully!"
