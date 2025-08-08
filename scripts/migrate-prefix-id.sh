#!/bin/bash
# Script to safely migrate prefix_id from schedules to courses table
# Includes safety checks and rollback capability

set -e

# Database configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../.env"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to execute MySQL commands safely
mysql_exec() {
    mysql -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" "$@" 2>/dev/null
}

echo "🔄 Migrating prefix_id from schedules to courses"
echo "==============================================="
echo "Database: $DB_NAME"
echo ""

# Safety check - verify database connection
info "Verifying database connection..."
if ! echo "SELECT 1;" | mysql_exec >/dev/null 2>&1; then
    error "Cannot connect to database $DB_NAME. Please check your credentials."
    exit 1
fi

# Check if both tables exist and have the expected columns
info "Checking table structure..."
schedules_has_prefix=$(mysql_exec -e "SHOW COLUMNS FROM schedules LIKE 'prefix_id';" | wc -l)
courses_has_prefix=$(mysql_exec -e "SHOW COLUMNS FROM courses LIKE 'prefix_id';" | wc -l)

if [[ "$schedules_has_prefix" == "0" ]]; then
    error "Schedules table does not have prefix_id column. Migration may have already been completed."
    exit 1
fi

if [[ "$courses_has_prefix" == "0" ]]; then
    error "Courses table does not have prefix_id column. Please run add-prefix-id-to-courses.sh first."
    exit 1
fi

# Count records before migration
info "Gathering pre-migration statistics..."
schedules_count=$(mysql_exec -e "SELECT COUNT(*) FROM schedules WHERE prefix_id IS NOT NULL;" -N)
courses_count=$(mysql_exec -e "SELECT COUNT(*) FROM courses;" -N)
courses_with_prefix_count=$(mysql_exec -e "SELECT COUNT(*) FROM courses WHERE prefix_id IS NOT NULL;" -N)

echo "Pre-migration state:"
echo "  - Schedules with prefix_id: $schedules_count"
echo "  - Total courses: $courses_count"
echo "  - Courses with prefix_id: $courses_with_prefix_count"
echo ""

# Confirm before proceeding
warning "This migration will:"
echo "  1. Copy prefix_id from schedules to all associated courses"
echo "  2. Remove the prefix_id column from the schedules table"
echo "  3. Update the unique constraint on schedules table"
echo ""
read -p "Continue with migration? (y/n): " confirm
if [[ "$confirm" != "y" ]]; then
    info "Migration cancelled."
    exit 0
fi

# Create backup
info "Creating backup of current state..."
backup_file="/tmp/schedules_backup_$(date +%Y%m%d_%H%M%S).sql"
mysqldump -h "$DB_HOST" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" schedules courses > "$backup_file"
info "Backup created: $backup_file"

# Perform the migration
info "Starting migration..."
mysql_exec < "$SCRIPT_DIR/../sql/migrate_prefix_id_from_schedules_to_courses.sql"

# Verify migration success
info "Verifying migration results..."
final_courses_with_prefix=$(mysql_exec -e "SELECT COUNT(*) FROM courses WHERE prefix_id IS NOT NULL;" -N)
schedules_has_prefix_after=$(mysql_exec -e "SHOW COLUMNS FROM schedules LIKE 'prefix_id';" | wc -l)

echo "Post-migration state:"
echo "  - Courses with prefix_id: $final_courses_with_prefix"
echo "  - Schedules table has prefix_id column: $([[ "$schedules_has_prefix_after" == "0" ]] && echo "No" || echo "Yes")"

if [[ "$schedules_has_prefix_after" == "0" ]] && [[ "$final_courses_with_prefix" -gt "$courses_with_prefix_count" ]]; then
    info "✅ Migration completed successfully!"
    echo ""
    echo "Summary:"
    echo "  - Migrated prefix_id to $((final_courses_with_prefix - courses_with_prefix_count)) additional courses"
    echo "  - Removed prefix_id column from schedules table"
    echo "  - Backup available at: $backup_file"
else
    error "❌ Migration may not have completed successfully."
    echo "Please check the database state manually."
    exit 1
fi

echo ""
echo "To rollback (if needed):"
echo "  mysql -h $DB_HOST -u $DB_USER -p'$DB_PASSWORD' $DB_NAME < $backup_file"
