#!/bin/bash
# Script to create prerequisites table in wmu_schedules_dev database

set -e

# Database configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../.env"
SQL_FILE="$SCRIPT_DIR/../sql/create_prerequisites_table.sql"

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

echo "📚 Creating Prerequisites Table"
echo "==============================="
echo "Database: $DB_NAME"
echo ""

# Check if database connection works
info "Verifying database connection..."
if ! echo "SELECT 1;" | mysql_exec >/dev/null 2>&1; then
    error "Cannot connect to database $DB_NAME. Please check your credentials."
    exit 1
fi

# Check if prefixes table exists (prerequisite for foreign keys)
info "Checking if prefixes table exists..."
if ! mysql_exec -e "SHOW TABLES LIKE 'prefixes';" | grep -q prefixes; then
    error "Prefixes table does not exist. Please create it first."
    exit 1
fi

# Check if prerequisites table already exists
info "Checking if prerequisites table already exists..."
if mysql_exec -e "SHOW TABLES LIKE 'prerequisites';" | grep -q prerequisites; then
    warning "Prerequisites table already exists."
    read -p "Do you want to drop and recreate it? (y/n): " recreate
    if [[ "$recreate" == "y" ]]; then
        info "Dropping existing prerequisites table..."
        mysql_exec -e "DROP TABLE prerequisites;"
    else
        info "Operation cancelled."
        exit 0
    fi
fi

# Create the prerequisites table
info "Creating prerequisites table..."
if mysql_exec < "$SQL_FILE"; then
    info "✅ Prerequisites table created successfully!"
    
    # Show table structure
    info "Table structure:"
    mysql_exec -e "DESCRIBE prerequisites;"
    
    # Show indexes
    info "Created indexes:"
    mysql_exec -e "SHOW INDEX FROM prerequisites WHERE Key_name != 'PRIMARY';"
    
    # Show foreign key constraints
    info "Foreign key constraints:"
    mysql_exec -e "SELECT 
        CONSTRAINT_NAME,
        COLUMN_NAME,
        REFERENCED_TABLE_NAME,
        REFERENCED_COLUMN_NAME
    FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
    WHERE TABLE_SCHEMA = '$DB_NAME' 
    AND TABLE_NAME = 'prerequisites' 
    AND REFERENCED_TABLE_NAME IS NOT NULL;"
    
else
    error "Failed to create prerequisites table."
    exit 1
fi

echo ""
echo "🎉 Prerequisites Table Setup Complete!"
echo "====================================="
echo ""
echo "Table structure:"
echo "  - id: Primary key"
echo "  - pred_prefix_id: Prerequisite course prefix (FK to prefixes)"
echo "  - pred_course_num: Prerequisite course number"
echo "  - succ_prefix_id: Successor course prefix (FK to prefixes)"
echo "  - succ_course_num: Successor course number"
echo ""
echo "Key pairs:"
echo "  - (pred_prefix_id, pred_course_num): Identifies prerequisite course"
echo "  - (succ_prefix_id, succ_course_num): Identifies successor course"
echo ""
echo "Example usage:"
echo "  -- Add CS 1120 as prerequisite for CS 2240"
echo "  INSERT INTO prerequisites (pred_prefix_id, pred_course_num, succ_prefix_id, succ_course_num)"
echo "  VALUES ("
echo "    (SELECT id FROM prefixes WHERE prefix = 'CS'),"
echo "    1120,"
echo "    (SELECT id FROM prefixes WHERE prefix = 'CS'),"
echo "    2240"
echo "  );"
