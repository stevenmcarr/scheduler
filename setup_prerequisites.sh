#!/bin/bash

# Prerequisites Database Setup Script
# This script creates the prerequisites table and sample data

echo "Setting up prerequisites database tables..."

# Note: Update these database connection details as needed
DB_HOST=${DB_HOST:-127.0.0.1}
DB_PORT=${DB_PORT:-3306}
DB_NAME=${DB_NAME:-wmu_schedules_dev}
DB_USER=${DB_USER:-root}

echo "Connecting to database: $DB_NAME at $DB_HOST:$DB_PORT"

# Check if mysql command is available
if ! command -v mysql &> /dev/null; then
    echo "Error: mysql command not found. Please install MySQL client."
    exit 1
fi

# Run the SQL script
echo "Creating prerequisites tables with foreign key constraints..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < create_prerequisites_table.sql

if [ $? -eq 0 ]; then
    echo "Prerequisites tables created successfully!"
    echo ""
    echo "The following tables have been created/updated:"
    echo "- prerequisites: Stores prerequisite relationships between courses"
    echo "- prefixes: Stores course prefixes for dropdown menus"
    echo ""
    echo "Sample data has been inserted for testing."
else
    echo "Error: Failed to create prerequisites tables."
    exit 1
fi
