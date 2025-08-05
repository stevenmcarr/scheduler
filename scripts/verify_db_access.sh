#!/bin/bash

# Script to verify database access for wmu_cs user
# This script tests various database operations to ensure proper access

set -e

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Database connection parameters
DB_USER="wmu_cs"
DB_PASSWORD="1h0ck3y$"
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_NAME="wmu_schedules_dev"

print_info "Testing database access for user: $DB_USER"
print_info "Database: $DB_NAME"
print_info "Host: $DB_HOST:$DB_PORT"
echo

# Test 1: Basic connection
print_info "Test 1: Basic database connection..."
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Database connection successful"
else
    print_error "Database connection failed"
    exit 1
fi

# Test 2: Database selection
print_info "Test 2: Database selection..."
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME;" > /dev/null 2>&1; then
    print_success "Database selection successful"
else
    print_error "Database selection failed"
    exit 1
fi

# Test 3: Table listing
print_info "Test 3: Table access..."
TABLES=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SHOW TABLES;" "$DB_NAME" --silent --skip-column-names 2>/dev/null | wc -l)
if [ "$TABLES" -gt 0 ]; then
    print_success "Can access $TABLES tables"
else
    print_error "Cannot access tables"
    exit 1
fi

# Test 4: Read operations
print_info "Test 4: Read operations..."
COURSE_COUNT=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT COUNT(*) FROM courses;" "$DB_NAME" --silent --skip-column-names 2>/dev/null)
if [ "$COURSE_COUNT" -ge 0 ]; then
    print_success "Can read data (found $COURSE_COUNT courses)"
else
    print_error "Cannot read data"
    exit 1
fi

# Test 5: Write operations (insert/update/delete test)
print_info "Test 5: Write operations..."
# Create a test table
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
CREATE TABLE IF NOT EXISTS test_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    test_data VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Can create tables"
else
    print_error "Cannot create tables"
    exit 1
fi

# Insert test data
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
INSERT INTO test_permissions (test_data) VALUES ('permission_test');" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Can insert data"
else
    print_error "Cannot insert data"
    exit 1
fi

# Update test data
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
UPDATE test_permissions SET test_data = 'permission_test_updated' WHERE test_data = 'permission_test';" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Can update data"
else
    print_error "Cannot update data"
    exit 1
fi

# Delete test data
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
DELETE FROM test_permissions WHERE test_data = 'permission_test_updated';" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Can delete data"
else
    print_error "Cannot delete data"
    exit 1
fi

# Clean up test table
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "DROP TABLE test_permissions;" "$DB_NAME" > /dev/null 2>&1
print_success "Test table cleaned up"

# Test 6: Schema operations
print_info "Test 6: Schema operations..."
if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
DESCRIBE courses;" "$DB_NAME" > /dev/null 2>&1; then
    print_success "Can describe table structure"
else
    print_error "Cannot describe table structure"
    exit 1
fi

echo
print_success "All database access tests passed!"
print_info "User '$DB_USER' has full access to database '$DB_NAME'"

# Show current database statistics
echo
print_info "Database Statistics:"
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "
SELECT 
    'Courses' as table_name, COUNT(*) as count FROM courses
UNION ALL
SELECT 
    'Instructors' as table_name, COUNT(*) as count FROM instructors
UNION ALL
SELECT 
    'Rooms' as table_name, COUNT(*) as count FROM rooms
UNION ALL
SELECT 
    'Schedules' as table_name, COUNT(*) as count FROM schedules
UNION ALL
SELECT 
    'Time Slots' as table_name, COUNT(*) as count FROM time_slots
UNION ALL
SELECT 
    'Users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 
    'Departments' as table_name, COUNT(*) as count FROM departments
UNION ALL
SELECT 
    'Prefixes' as table_name, COUNT(*) as count FROM prefixes;" "$DB_NAME" 2>/dev/null
