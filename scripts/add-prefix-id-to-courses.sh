#!/bin/bash
# Script to add prefix_id column to courses table in wmu_schedules_dev database
# This column will be a foreign key reference to the prefixes table

set -e

# Database configuration
DB_NAME="wmu_schedules_dev"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="$SCRIPT_DIR/../sql/add_prefix_id_to_courses.sql"

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

# Check if .env file exists
if [[ ! -f "$SCRIPT_DIR/../.env" ]]; then
    error ".env file not found. Please create it with database credentials."
    exit 1
fi

# Source environment variables
source "$SCRIPT_DIR/../.env"

# Validate required environment variables
if [[ -z "$DB_USER" || -z "$DB_PASSWORD" || -z "$DB_HOST" ]]; then
    error "Missing required environment variables: DB_USER, DB_PASSWORD, DB_HOST"
    exit 1
fi

# Function to execute MySQL commands safely
mysql_exec() {
    local mysql_args=()
    mysql_args+=("-h" "$DB_HOST")
    mysql_args+=("-u" "$DB_USER")
    mysql_args+=("-p$DB_PASSWORD")
    mysql_args+=("$DB_NAME")
    
    mysql "${mysql_args[@]}" "$@" 2>/dev/null
}

# Check if database exists
info "Checking if database $DB_NAME exists..."
if ! mysql_exec -e "SELECT 1;" >/dev/null 2>&1; then
    error "Cannot connect to database $DB_NAME. Please check your credentials and database."
    exit 1
fi

# Check if courses table exists
info "Checking if courses table exists..."
if ! mysql_exec -e "SHOW TABLES LIKE 'courses';" | grep -q courses; then
    error "Courses table does not exist in database $DB_NAME."
    exit 1
fi

# Check if prefixes table exists
info "Checking if prefixes table exists..."
if ! mysql_exec -e "SHOW TABLES LIKE 'prefixes';" | grep -q prefixes; then
    error "Prefixes table does not exist in database $DB_NAME."
    exit 1
fi

# Check if prefix_id column already exists
info "Checking if prefix_id column already exists..."
if mysql_exec -e "SHOW COLUMNS FROM courses LIKE 'prefix_id';" | grep -q prefix_id; then
    warning "Column prefix_id already exists in courses table."
    read -p "Do you want to recreate it? (y/n): " recreate
    if [[ "$recreate" != "y" ]]; then
        info "Operation cancelled."
        exit 0
    fi
    info "Dropping existing prefix_id column..."
    mysql_exec -e "ALTER TABLE courses DROP FOREIGN KEY IF EXISTS fk_courses_prefix_id;" || true
    mysql_exec -e "ALTER TABLE courses DROP INDEX IF EXISTS idx_prefix_id;" || true
    mysql_exec -e "ALTER TABLE courses DROP COLUMN IF EXISTS prefix_id;" || true
fi

# Create the SQL file
info "Creating SQL file: $SQL_FILE"
cat > "$SQL_FILE" << 'EOF'
-- Add prefix_id column to courses table
-- This establishes a relationship between courses and their subject prefixes

USE wmu_schedules_dev;

-- Add the prefix_id column
ALTER TABLE courses 
ADD COLUMN prefix_id INT NULL AFTER schedule_id;

-- Add index for performance
ALTER TABLE courses 
ADD INDEX idx_prefix_id (prefix_id);

-- Add foreign key constraint
ALTER TABLE courses 
ADD CONSTRAINT fk_courses_prefix_id 
FOREIGN KEY (prefix_id) REFERENCES prefixes(id) ON DELETE SET NULL;

-- Show the updated table structure
DESCRIBE courses;
EOF

# Execute the SQL file
info "Adding prefix_id column to courses table..."
if mysql_exec < "$SQL_FILE"; then
    info "✅ Successfully added prefix_id column to courses table!"
    
    # Verify the change
    info "Verifying the new column..."
    mysql_exec -e "SHOW COLUMNS FROM courses WHERE Field = 'prefix_id';"
    
    info "Column details:"
    mysql_exec -e "SELECT 
        COLUMN_NAME,
        DATA_TYPE,
        IS_NULLABLE,
        COLUMN_DEFAULT,
        EXTRA
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = '$DB_NAME' 
    AND TABLE_NAME = 'courses' 
    AND COLUMN_NAME = 'prefix_id';"
    
    # Show foreign key constraint
    info "Foreign key constraint:"
    mysql_exec -e "SELECT 
        CONSTRAINT_NAME,
        COLUMN_NAME,
        REFERENCED_TABLE_NAME,
        REFERENCED_COLUMN_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
    WHERE TABLE_SCHEMA = '$DB_NAME' 
    AND TABLE_NAME = 'courses' 
    AND COLUMN_NAME = 'prefix_id';"
    
else
    error "Failed to add prefix_id column to courses table."
    exit 1
fi

info "Script completed successfully!"
info ""
info "Next steps:"
info "1. Update your application code to populate prefix_id when creating/updating courses"
info "2. Consider migrating existing course data to set appropriate prefix_id values"
info "3. Update your course creation forms and validation logic"
