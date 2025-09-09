#!/bin/bash

# Script to add department_id column to users table
# This script will run the migration for both dev and production databases

echo "Adding department_id column to users table..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Error: .env file not found. Please make sure you're in the correct directory."
    exit 1
fi

# Source the environment variables
source .env

# Run migration for development database (wmu_schedules_dev)
echo "Running migration for wmu_schedules_dev..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" wmu_schedules_dev < sql/add_department_to_users.sql

if [ $? -eq 0 ]; then
    echo "✅ Successfully added department_id column to wmu_schedules_dev.users table"
else
    echo "❌ Failed to add department_id column to wmu_schedules_dev.users table"
    exit 1
fi

# Run migration for production database (wmu_schedules)
echo "Running migration for wmu_schedules..."
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" wmu_schedules < sql/add_department_to_users.sql

if [ $? -eq 0 ]; then
    echo "✅ Successfully added department_id column to wmu_schedules.users table"
else
    echo "❌ Failed to add department_id column to wmu_schedules.users table"
    exit 1
fi

echo ""
echo "Migration completed successfully!"
echo "The users table now has a department_id column that references the departments table."
echo ""
echo "Next steps:"
echo "1. Test the application to ensure everything works correctly"
echo "2. Update existing users to assign them to appropriate departments if needed"
