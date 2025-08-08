#!/bin/bash
# Script to grant wmu_cs user full access rights to wmu_schedules_dev database

set -e

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

# Database configuration
TARGET_DB="wmu_schedules_dev"
DEV_USER="wmu_cs"
DEV_PASSWORD="1h0ck3y\$"

echo "🔐 Granting Database Access Rights"
echo "=================================="
echo "Database: $TARGET_DB"
echo "User: $DEV_USER"
echo ""

# Check if we can connect as root
info "Checking MySQL root access..."
read -s -p "Enter MySQL root password: " MYSQL_ROOT_PASSWORD
echo ""

# Function to execute MySQL commands as root
mysql_root_exec() {
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" "$@" 2>/dev/null
}

# Verify root password works
info "Verifying root password..."
if ! echo "SELECT 1;" | mysql_root_exec >/dev/null 2>&1; then
    error "Invalid root password or connection failed."
    exit 1
fi

# Check if target database exists
info "Checking if database '$TARGET_DB' exists..."
if ! echo "USE $TARGET_DB;" | mysql_root_exec >/dev/null 2>&1; then
    warning "Database '$TARGET_DB' does not exist. Creating it..."
    mysql_root_exec -e "CREATE DATABASE $TARGET_DB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    info "✅ Database '$TARGET_DB' created successfully."
fi

# Check if user exists, create if not
info "Checking if user '$DEV_USER' exists..."
user_exists=$(mysql_root_exec -e "SELECT COUNT(*) FROM mysql.user WHERE user='$DEV_USER' AND host='localhost';" -N 2>/dev/null || echo "0")

if [[ "$user_exists" == "0" ]]; then
    info "Creating user '$DEV_USER'..."
    mysql_root_exec -e "CREATE USER '$DEV_USER'@'localhost' IDENTIFIED BY '$DEV_PASSWORD';"
    info "✅ User '$DEV_USER' created successfully."
else
    info "User '$DEV_USER' already exists."
    # Update password to ensure it matches
    mysql_root_exec -e "ALTER USER '$DEV_USER'@'localhost' IDENTIFIED BY '$DEV_PASSWORD';"
    info "✅ User password updated."
fi

# Grant full privileges to the database
info "Granting ALL PRIVILEGES on '$TARGET_DB' to '$DEV_USER'..."
mysql_root_exec -e "GRANT ALL PRIVILEGES ON $TARGET_DB.* TO '$DEV_USER'@'localhost';"

# Also grant some global privileges that might be needed
info "Granting additional privileges..."
mysql_root_exec -e "GRANT CREATE, DROP, INDEX, ALTER ON *.* TO '$DEV_USER'@'localhost';"

# Flush privileges
mysql_root_exec -e "FLUSH PRIVILEGES;"

# Verify the setup
info "Verifying access rights..."

# Test database connection as development user
if echo "USE $TARGET_DB; SHOW TABLES;" | mysql -u "$DEV_USER" -p"$DEV_PASSWORD" >/dev/null 2>&1; then
    info "✅ User '$DEV_USER' can access database '$TARGET_DB' successfully!"
else
    error "❌ User '$DEV_USER' cannot access database '$TARGET_DB'."
    exit 1
fi

# Show granted privileges
info "📋 Current privileges for '$DEV_USER'@'localhost':"
mysql_root_exec -e "SHOW GRANTS FOR '$DEV_USER'@'localhost';"

# Test creating a table to verify write permissions
info "Testing write permissions..."
test_result=$(mysql -u "$DEV_USER" -p"$DEV_PASSWORD" "$TARGET_DB" -e "
CREATE TABLE IF NOT EXISTS test_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    test_column VARCHAR(50)
);
INSERT INTO test_permissions (test_column) VALUES ('access_test');
SELECT COUNT(*) FROM test_permissions WHERE test_column = 'access_test';
DROP TABLE test_permissions;
" -N 2>/dev/null || echo "0")

if [[ "$test_result" == "1" ]]; then
    info "✅ Write permissions verified successfully!"
else
    error "❌ Write permissions test failed."
    exit 1
fi

echo ""
echo "🎉 Database Access Rights Granted Successfully!"
echo "=============================================="
echo "Database: $TARGET_DB"
echo "User: $DEV_USER"
echo "Privileges: ALL PRIVILEGES on $TARGET_DB.*"
echo ""
echo "Connection test command:"
echo "mysql -u $DEV_USER -p'$DEV_PASSWORD' $TARGET_DB"
echo ""
echo "Your .env configuration is correct:"
echo "DB_USER=$DEV_USER"
echo "DB_PASSWORD=$DEV_PASSWORD"
echo "DB_NAME=$TARGET_DB"
