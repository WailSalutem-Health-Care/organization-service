#!/bin/bash
#
# One-time database migration runner for organization-service
# Applies SQL migrations in order and tracks them in wailsalutem.schema_migrations
#
# Usage:
#   ./scripts/run_migrations.sh --deployed
#
# IMPORTANT:
# - This script is TEMPORARY
# - Run it ONLY ONCE against an EMPTY deployed database
# - Do NOT use this for normal deployments
#

set -euo pipefail

echo "🚀 Starting database migrations for organization-service"

# ----------- ENV SELECTION -----------

if [ "${1:-}" != "--deployed" ]; then
  echo "❌ ERROR: You must run this script with --deployed"
  echo "Usage: ./scripts/run_migrations.sh --deployed"
  exit 1
fi

# ----------- REQUIRED ENV VARS -----------

: "${DEPLOYED_DB_HOST:?DEPLOYED_DB_HOST is required}"
: "${DEPLOYED_DB_USER:?DEPLOYED_DB_USER is required}"
: "${DEPLOYED_DB_PASSWORD:?DEPLOYED_DB_PASSWORD is required}"
: "${DEPLOYED_DB_NAME:?DEPLOYED_DB_NAME is required}"

DB_HOST="$DEPLOYED_DB_HOST"
DB_PORT="${DEPLOYED_DB_PORT:-5432}"
DB_USER="$DEPLOYED_DB_USER"
DB_PASSWORD="$DEPLOYED_DB_PASSWORD"
DB_NAME="$DEPLOYED_DB_NAME"

export PGPASSWORD="$DB_PASSWORD"

echo "📍 Target database: $DB_HOST:$DB_PORT/$DB_NAME"
echo ""

# ----------- VERIFY CONNECTION -----------

echo "🔍 Checking database connectivity..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null
echo "✅ Database connection OK"
echo ""

# ----------- MIGRATION TRACKING TABLE -----------

echo "🔧 Ensuring migration tracking table exists..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<'EOF'
CREATE SCHEMA IF NOT EXISTS wailsalutem;

CREATE TABLE IF NOT EXISTS wailsalutem.schema_migrations (
    version        VARCHAR(50) PRIMARY KEY,
    description    TEXT,
    applied_at     TIMESTAMP DEFAULT now()
);
EOF

echo "✅ Migration tracking ready"
echo ""

# ----------- APPLY MIGRATIONS -----------

MIGRATION_DIR="$(dirname "$0")/../migrations"

if [ ! -d "$MIGRATION_DIR" ]; then
  echo "❌ ERROR: migrations directory not found at $MIGRATION_DIR"
  exit 1
fi

echo "📂 Using migrations directory: $MIGRATION_DIR"
echo ""

for sql_file in $(ls "$MIGRATION_DIR"/V*.sql | sort); do
  filename=$(basename "$sql_file")
  version="${filename%%__*}"
  description="${filename#*__}"
  description="${description%.sql}"
  description="${description//_/ }"

  already_applied=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc \
    "SELECT COUNT(*) FROM wailsalutem.schema_migrations WHERE version = '$version'")

  if [ "$already_applied" -gt 0 ]; then
    echo "⏭️  $version already applied"
    continue
  fi

  echo "📄 Applying $version — $description"

  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file"

  psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
    "INSERT INTO wailsalutem.schema_migrations (version, description)
     VALUES ('$version', '$description')"

  echo "✅ $version applied"
  echo ""
done

unset PGPASSWORD

echo "🎉 All migrations applied successfully"
echo ""
echo "📊 Applied migrations:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
  "SELECT version, description, applied_at FROM wailsalutem.schema_migrations ORDER BY version;"
