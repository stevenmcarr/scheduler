#!/bin/bash
# Migration script runner to move prefix from schedules to courses table
# Updated with better error handling

set -e

echo "🔄 Starting database migration: Move prefix from schedules to courses table"

# Check if .env file exists and source it
if [ -f ".env" ]; then
    source .env
    echo "✅ Loaded database configuration from .env"
else
    echo "❌ Error: .env file not found"
    exit 1
fi

# Validate required environment variables
if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
    echo "❌ Error: Missing required database configuration in .env file"
    echo "Required variables: DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
    exit 1
fi

# Create migrations directory if it doesn't exist
mkdir -p scripts/migrations

# Check if migration file exists
MIGRATION_FILE="sql/migrate_prefix_to_courses.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Error: Migration file not found: $MIGRATION_FILE"
    exit 1
fi

# Create backup directory if it doesn't exist
mkdir -p backups

# Test database connection
echo "🔍 Testing database connection..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD -e "SELECT 1" $DB_NAME > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Database connection successful"
else
    echo "❌ Failed to connect to database. Please check your credentials."
    exit 1
fi

# Run pre-migration check first
echo "🔍 Running pre-migration check..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < sql/check_migration.sql

# Create backup before migration
echo "📁 Creating backup before migration..."
BACKUP_FILE="backups/before_prefix_migration_$(date +%Y%m%d_%H%M%S).sql"
mysqldump -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME > $BACKUP_FILE

if [ $? -eq 0 ]; then
    echo "✅ Backup created successfully: $BACKUP_FILE"
else
    echo "❌ Failed to create backup. Aborting migration."
    exit 1
fi

# Confirm before proceeding
echo ""
echo "⚠️  WARNING: This migration will modify the database structure."
echo "📋 Changes to be made:"
echo "   1. Add 'prefix' column to courses table (if not exists)"
echo "   2. Copy prefix data from schedules -> prefixes -> courses"
echo "   3. Remove prefix_id column from schedules table (if exists)"
echo "   4. Drop foreign key constraint from schedules to prefixes (if exists)"
echo ""
echo "📄 Backup location: $BACKUP_FILE"
echo ""
read -p "Do you want to proceed with the migration? (y/N): " confirm

if [[ $confirm != [yY] && $confirm != [yY][eE][sS] ]]; then
    echo "❌ Migration cancelled by user"
    exit 0
fi

# Run the migration
echo ""
echo "🚀 Running migration..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < $MIGRATION_FILE

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Migration completed successfully!"
    echo "📊 Running verification..."
    
    # Run verification script
    if [ -f "scripts/verify_migration.sh" ]; then
        ./scripts/verify_migration.sh
    fi
    
    echo ""
    echo "📄 Backup available at: $BACKUP_FILE"
    echo ""
    echo "🔍 Manual verification commands:"
    echo "   mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME"
    echo "   mysql> SELECT COUNT(*) FROM courses WHERE prefix != '';"
    echo "   mysql> DESCRIBE courses;"
    echo "   mysql> DESCRIBE schedules;"
else
    echo ""
    echo "❌ Migration failed!"
    echo "📄 Database backup is available at: $BACKUP_FILE"
    echo "🔧 You can restore the backup using:"
    echo "   mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < $BACKUP_FILE"
    exit 1
fi