#!/bin/bash

# Script to create a development database that is a copy of the production wmu_schedules database
# This script will:
# 1. Create a new database called wmu_schedules_dev
# 2. Copy the structure and data from wmu_schedules to wmu_schedules_dev
# 3. Optionally clean up sensitive data for development use

set -e  # Exit on any error

# Get the project root directory (parent of scripts directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Default values
PROD_DB="wmu_schedules"
DEV_DB="wmu_schedules_dev"
BACKUP_FILE="/tmp/wmu_schedules_backup_$(date +%Y%m%d_%H%M%S).sql"
CLEAN_SENSITIVE_DATA=false
DROP_EXISTING=false

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -u, --user USER         MySQL username (required)"
    echo "  -p, --password PASS     MySQL password (optional - will prompt securely if not provided)"
    echo "  -H, --host HOST         MySQL host (default: localhost)"
    echo "  -P, --port PORT         MySQL port (default: 3306)"
    echo "  -s, --source-db DB      Source database name (default: wmu_schedules)"
    echo "  -d, --dev-db DB         Development database name (default: wmu_schedules_dev)"
    echo "  -c, --clean-sensitive   Clean sensitive data from development database"
    echo "  -f, --force             Drop existing development database if it exists"
    echo "  --backup-file FILE      Custom backup file path (default: /tmp/wmu_schedules_backup_TIMESTAMP.sql)"
    echo ""
    echo "Examples:"
    echo "  $0 -u root                           # Will prompt for password securely"
    echo "  $0 -u root -p mypassword             # Password on command line (less secure)"
    echo "  $0 -u dbuser -H 192.168.1.100 -c -f # Will prompt for password"
    echo "  $0 -u dbuser --clean-sensitive --force"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -u|--user)
            DB_USER="$2"
            shift 2
            ;;
        -p|--password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        -H|--host)
            DB_HOST="$2"
            shift 2
            ;;
        -P|--port)
            DB_PORT="$2"
            shift 2
            ;;
        -s|--source-db)
            PROD_DB="$2"
            shift 2
            ;;
        -d|--dev-db)
            DEV_DB="$2"
            shift 2
            ;;
        -c|--clean-sensitive)
            CLEAN_SENSITIVE_DATA=true
            shift
            ;;
        -f|--force)
            DROP_EXISTING=true
            shift
            ;;
        --backup-file)
            BACKUP_FILE="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Set default values if not provided
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}

# Check if required parameters are provided
if [[ -z "$DB_USER" ]]; then
    print_error "MySQL username is required. Use -u or --user option."
    show_usage
    exit 1
fi

# Construct MySQL connection arguments
# Use -p without password to prompt for password securely
if [[ -n "$DB_PASSWORD" ]]; then
    # Password provided via command line (less secure but still supported)
    MYSQL_CMD_ARGS=(-h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD")
    MYSQLDUMP_CMD_ARGS=(-h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD")
    PASSWORD_MODE="provided"
else
    # No password provided - use -p to prompt for password (more secure)
    MYSQL_CMD_ARGS=(-h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p)
    MYSQLDUMP_CMD_ARGS=(-h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p)
    PASSWORD_MODE="prompt"
fi

# Create functions to execute MySQL commands safely
mysql_exec() {
    if [[ "$PASSWORD_MODE" == "prompt" ]]; then
        echo "MySQL password required for: $*" >&2
    fi
    mysql "${MYSQL_CMD_ARGS[@]}" "$@"
}

mysqldump_exec() {
    if [[ "$PASSWORD_MODE" == "prompt" ]]; then
        echo "MySQL password required for mysqldump operation" >&2
    fi
    mysqldump "${MYSQLDUMP_CMD_ARGS[@]}" "$@"
}

print_status "Starting development database creation process..."
print_status "Source database: $PROD_DB"
print_status "Development database: $DEV_DB"
print_status "MySQL host: $DB_HOST:$DB_PORT"
print_status "MySQL user: $DB_USER"
if [[ "$PASSWORD_MODE" == "prompt" ]]; then
    print_status "Password mode: Will prompt for password at each MySQL operation (secure)"
else
    print_status "Password mode: Using provided password (less secure)"
fi

# Test MySQL connection
print_status "Testing MySQL connection..."
if ! mysql_exec -e "SELECT 1;" > /dev/null 2>&1; then
    print_error "Failed to connect to MySQL server. Please check your credentials and connection settings."
    exit 1
fi
print_success "MySQL connection successful"

# Check if source database exists
print_status "Checking if source database '$PROD_DB' exists..."
if ! mysql_exec -e "USE $PROD_DB;" > /dev/null 2>&1; then
    print_error "Source database '$PROD_DB' does not exist."
    exit 1
fi
print_success "Source database '$PROD_DB' found"

# Check if development database already exists
print_status "Checking if development database '$DEV_DB' exists..."
if mysql_exec -e "USE $DEV_DB;" > /dev/null 2>&1; then
    if [[ "$DROP_EXISTING" == true ]]; then
        print_warning "Development database '$DEV_DB' already exists. Dropping it..."
        mysql_exec -e "DROP DATABASE $DEV_DB;"
        print_success "Existing development database dropped"
    else
        print_error "Development database '$DEV_DB' already exists. Use -f or --force to drop it."
        exit 1
    fi
fi

# Create development database
print_status "Creating development database '$DEV_DB'..."
mysql_exec -e "CREATE DATABASE $DEV_DB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
print_success "Development database '$DEV_DB' created"

# Create backup of source database
print_status "Creating backup of source database to '$BACKUP_FILE'..."
mysqldump_exec --single-transaction --routines --triggers --databases $PROD_DB > "$BACKUP_FILE"
print_success "Backup created successfully"

# Copy structure and data to development database
print_status "Copying structure and data to development database..."
# Extract just the database content (without CREATE DATABASE statement)
sed "s/CREATE DATABASE \`$PROD_DB\`/CREATE DATABASE IF NOT EXISTS \`$DEV_DB\`/g; s/USE \`$PROD_DB\`/USE \`$DEV_DB\`/g" "$BACKUP_FILE" | mysql_exec
print_success "Data copied to development database"

# Clean sensitive data if requested
if [[ "$CLEAN_SENSITIVE_DATA" == true ]]; then
    print_status "Cleaning sensitive data from development database..."
    
    # Create SQL script for cleaning sensitive data
    cat > /tmp/clean_sensitive_data.sql << EOF
USE $DEV_DB;

-- Clean user passwords and reset to a development default
-- Note: Using actual column names from users table (password, administrator)
UPDATE users SET 
    password = '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: 'password'
    email = CONCAT('dev_user_', id, '@example.com')
WHERE administrator != 1;

-- Keep admin accounts but reset password
UPDATE users SET 
    password = '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi' -- password: 'password'
WHERE administrator = 1;

-- Add a default development admin user if not exists
INSERT IGNORE INTO users (username, password, administrator, email, created_at) 
VALUES ('dev_admin', '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 1, 'dev_admin@example.com', NOW());

-- Clean instructor personal information (optional)
-- UPDATE instructors SET 
--     email = CONCAT('instructor_', id, '@example.com'),
--     phone = '555-0100'
-- WHERE email IS NOT NULL;

-- Add development prefix to schedule names for clarity
UPDATE schedules SET 
    term = CONCAT('[DEV] ', term)
WHERE term NOT LIKE '[DEV]%';
EOF

    mysql_exec < /tmp/clean_sensitive_data.sql
    rm /tmp/clean_sensitive_data.sql
    
    print_success "Sensitive data cleaned from development database"
    print_warning "All user passwords have been reset to 'password' for development"
fi

# Create development-specific .env file
print_status "Creating development environment file..."
cat > "$PROJECT_ROOT/.env.dev" << EOF
# Development Database Configuration
# Copy of production database: $PROD_DB -> $DEV_DB
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DEV_DB

# Server Configuration
SERVER_PORT=8081

# TLS Configuration (disabled for development)
TLS_ENABLED=false
TLS_CERT_FILE=
TLS_KEY_FILE=

# Development Environment
ENVIRONMENT=development
DEBUG=true
EOF

print_success "Development environment file created at .env.dev"

# Verify the copy was successful
print_status "Verifying database copy..."
PROD_TABLE_COUNT=$(mysql_exec -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$PROD_DB';" -s -N)
DEV_TABLE_COUNT=$(mysql_exec -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$DEV_DB';" -s -N)

if [[ "$PROD_TABLE_COUNT" != "$DEV_TABLE_COUNT" ]]; then
    print_error "Table count mismatch! Production: $PROD_TABLE_COUNT, Development: $DEV_TABLE_COUNT"
    exit 1
fi

print_success "Database copy verification successful"
print_success "Production tables: $PROD_TABLE_COUNT, Development tables: $DEV_TABLE_COUNT"

# Show summary
echo
print_success "Development database setup completed successfully!"
echo
echo "Summary:"
echo "  • Source database: $PROD_DB"
echo "  • Development database: $DEV_DB"
echo "  • Backup file: $BACKUP_FILE"
echo "  • Environment file: .env.dev"
if [[ "$CLEAN_SENSITIVE_DATA" == true ]]; then
echo "  • Sensitive data cleaned"
echo "  • Default login: dev_admin / password"
fi
echo
echo "To use the development database:"
echo "  1. Set environment variable: export ENV_FILE=.env.dev"
echo "  2. Or run with: ENV_FILE=.env.dev ./bin/scheduler"
echo "  3. Development server will run on port 8081"
echo
print_warning "Remember to clean up the backup file when no longer needed:"
print_warning "  rm '$BACKUP_FILE'"
