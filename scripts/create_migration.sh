#!/bin/bash

# Script to create a new migration file
# Usage: ./scripts/create_migration.sh description_of_migration

if [ $# -eq 0 ]; then
    echo "Usage: $0 <migration_description>"
    echo "Example: $0 add_user_favorites"
    exit 1
fi

DESCRIPTION=$1
TIMESTAMP=$(date +%Y%m%d%H%M%S)
FILENAME="${TIMESTAMP}_${DESCRIPTION}.sql"
FILEPATH="storage/migrations/${FILENAME}"

# Create the migration file with template
cat > "$FILEPATH" << EOF
-- +goose Up
-- Add your SQL statements here for the up migration


-- +goose Down
-- Add your SQL statements here for the down migration

EOF

echo "Created migration file: $FILEPATH"
echo "Edit the file to add your SQL statements, then run:"
echo "  make migrate-up"
