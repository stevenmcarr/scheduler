#!/bin/bash
# Script to add 'AO' as a valid mode value to the courses table

set -e

# Database configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../.env"
SQL_FILE="$SCRIPT_DIR/../sql/add_ao_mode_to_courses.sql"

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

echo "🔧 Adding 'AO' Mode to Courses Table"
echo "===================================="
echo "Database: $DB_NAME"
echo ""

# Check if database connection works
info "Verifying database connection..."
if ! echo "SELECT 1;" | mysql_exec >/dev/null 2>&1; then
    error "Cannot connect to database $DB_NAME. Please check your credentials."
    exit 1
fi

# Check if courses table exists
info "Checking if courses table exists..."
if ! mysql_exec -e "SHOW TABLES LIKE 'courses';" | grep -q courses; then
    error "Courses table does not exist in database $DB_NAME."
    exit 1
fi

# Check current mode column definition
info "Checking current mode column definition..."
current_enum=$(mysql_exec -e "SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '$DB_NAME' AND TABLE_NAME = 'courses' AND COLUMN_NAME = 'mode';" -N)

if [[ -z "$current_enum" ]]; then
    error "Mode column not found in courses table."
    exit 1
fi

echo "Current mode definition: $current_enum"

# Check if 'AO' is already in the enum
if [[ "$current_enum" == *"'AO'"* ]]; then
    warning "'AO' is already a valid mode value."
    read -p "Continue anyway to verify? (y/n): " continue_anyway
    if [[ "$continue_anyway" != "y" ]]; then
        info "Operation cancelled."
        exit 0
    fi
fi

# Show current mode values in use
info "Current mode values in use:"
mysql_exec -e "SELECT mode, COUNT(*) as count FROM courses GROUP BY mode ORDER BY mode;"

# Confirm before proceeding
echo ""
warning "This will modify the mode column ENUM definition to include 'AO'."
echo "Current ENUM: $current_enum"
echo "New ENUM will be: enum('IP','FSO','PSO','H','CLAS','AO')"
echo ""
read -p "Continue with the modification? (y/n): " confirm
if [[ "$confirm" != "y" ]]; then
    info "Operation cancelled."
    exit 0
fi

# Execute the SQL script
info "Modifying mode column to include 'AO'..."
if mysql_exec < "$SQL_FILE"; then
    info "✅ Mode column updated successfully!"
    
    # Verify the change
    info "Verifying the change..."
    new_enum=$(mysql_exec -e "SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '$DB_NAME' AND TABLE_NAME = 'courses' AND COLUMN_NAME = 'mode';" -N)
    echo "Updated mode definition: $new_enum"
    
    # Show column details
    info "Mode column details:"
    mysql_exec -e "SHOW COLUMNS FROM courses LIKE 'mode';"
    
    if [[ "$new_enum" == *"'AO'"* ]]; then
        info "✅ 'AO' has been successfully added to the mode column!"
    else
        error "❌ 'AO' was not found in the updated column definition."
        exit 1
    fi
    
else
    error "Failed to update mode column."
    exit 1
fi

echo ""
echo "🎉 Mode Column Update Complete!"
echo "==============================="
echo "The courses table mode column now accepts these values:"
echo "  - IP (In Person)"
echo "  - FSO (Fully Synchronous Online)"  
echo "  - PSO (Partially Synchronous Online)"
echo "  - H (Hybrid)"
echo "  - CLAS (Classical)"
echo "  - AO (Asynchronous Online) ← NEW"
echo ""
echo "You can now insert or update courses with mode = 'AO'"
