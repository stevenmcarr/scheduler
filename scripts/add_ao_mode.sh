#!/bin/bash
# Script to add AO mode to the courses table

set -e

echo "🔄 Adding AO mode to courses table..."

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

# Check if migration file exists
MIGRATION_FILE="sql/add_ao_mode.sql"
if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Error: Migration file not found: $MIGRATION_FILE"
    exit 1
fi

# Test database connection
echo "🔍 Testing database connection..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD -e "SELECT 1" $DB_NAME > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Database connection successful"
else
    echo "❌ Failed to connect to database. Please check your credentials."
    exit 1
fi

# Create backup directory if it doesn't exist
mkdir -p backups

# Create backup before migration
echo "📁 Creating backup before adding AO mode..."
BACKUP_FILE="backups/before_add_ao_mode_$(date +%Y%m%d_%H%M%S).sql"
mysqldump -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME > $BACKUP_FILE

if [ $? -eq 0 ]; then
    echo "✅ Backup created successfully: $BACKUP_FILE"
else
    echo "❌ Failed to create backup. Aborting migration."
    exit 1
fi

# Run the migration
echo ""
echo "🚀 Adding AO mode to courses table..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < $MIGRATION_FILE

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ AO mode added successfully!"
    echo "📊 Summary of changes:"
    echo "   ✓ Added 'AO' (Asynchronous Online) to mode enum in courses table"
    echo "   ✓ Valid modes are now: IP, FSO, PSO, H, CLAS, AO"
    echo ""
    echo "📄 Backup available at: $BACKUP_FILE"
    echo ""
    echo "🔍 To verify the change:"
    echo "   mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME"
    echo "   mysql> SHOW COLUMNS FROM courses LIKE 'mode';"
else
    echo ""
    echo "❌ Migration failed!"
    echo "📄 Database backup is available at: $BACKUP_FILE"
    echo "🔧 You can restore the backup using:"
    echo "   mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < $BACKUP_FILE"
    exit 1
fi
