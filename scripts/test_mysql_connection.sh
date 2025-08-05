#!/bin/bash
# Test the fixed MySQL command construction

DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="root"
DB_PASSWORD='9h0c$$k3y%'

echo "=== Testing Fixed MySQL Command Construction ==="
echo "Password: $DB_PASSWORD"
echo

# Test the array-based approach
MYSQL_CMD_ARGS=(-h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD")

mysql_exec() {
    mysql "${MYSQL_CMD_ARGS[@]}" "$@"
}

echo "Array contents:"
printf '%s\n' "${MYSQL_CMD_ARGS[@]}"
echo

echo "Testing connection (this should work now):"
if mysql_exec -e "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ SUCCESS: MySQL connection works with special characters!"
else
    echo "❌ FAILED: Still having issues with MySQL connection"
    echo "Error details:"
    mysql_exec -e "SELECT 1;" 2>&1
fi
