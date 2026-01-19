#!/bin/bash
#
# Setup test database with migrations
# This script runs all migrations on the wailsalutem_test database
#

set -e

echo "🔧 Setting up test database..."

# Database connection details
DB_USER="localadmin"
DB_NAME="wailsalutem_test"
CONTAINER="wailsalutem-postgres"

# Check if container is running
if ! docker ps | grep -q "$CONTAINER"; then
    echo "❌ Error: PostgreSQL container '$CONTAINER' is not running"
    echo "Please start it with: cd ../WailSalutem-infra && docker-compose up -d postgres"
    exit 1
fi

echo "✅ PostgreSQL container is running"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATION_DIR="$SCRIPT_DIR/../migrations"

if [ ! -d "$MIGRATION_DIR" ]; then
    echo "❌ Error: migrations directory not found at $MIGRATION_DIR"
    exit 1
fi

echo "📂 Using migrations directory: $MIGRATION_DIR"
echo ""

# Apply each migration
for sql_file in "$MIGRATION_DIR"/V*.sql; do
    if [ -f "$sql_file" ]; then
        filename=$(basename "$sql_file")
        echo "📄 Applying $filename..."
        
        docker exec -i "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" < "$sql_file"
        
        if [ $? -eq 0 ]; then
            echo "✅ $filename applied successfully"
        else
            echo "❌ Failed to apply $filename"
            exit 1
        fi
        echo ""
    fi
done

echo "🎉 Test database setup complete!"
echo ""
echo "📊 Verifying schema..."
docker exec "$CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "\dt wailsalutem.*"
