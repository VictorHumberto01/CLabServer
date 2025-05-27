#!/bin/bash

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql is not installed"
    exit 1
fi

# Check if database exists
if ! psql -lqt | cut -d \| -f 1 | grep -qw clab; then
    echo "Creating database clab..."
    createdb clab
fi

# Run migrations
echo "Running migrations..."
for file in internal/database/migrations/*.up.sql; do
    echo "Running migration: $file"
    psql -d clab -f "$file"
done

echo "Migrations completed successfully!" 