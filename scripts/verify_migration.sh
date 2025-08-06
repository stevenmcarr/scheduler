#!/bin/bash
# Quick verification script to check migration results

set -e

# Check if .env file exists and source it
if [ -f ".env" ]; then
    source .env
else
    echo "❌ Error: .env file not found"
    exit 1
fi

echo "🔍 Verifying database migration results..."
echo ""

# Check courses table structure
echo "📋 Courses table structure:"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME -e "DESCRIBE courses;" | grep -E "(Field|prefix)"

echo ""

# Check schedules table structure  
echo "📋 Schedules table structure:"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME -e "DESCRIBE schedules;"

echo ""

# Count courses with prefix data
echo "📊 Courses with prefix data:"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME -e "SELECT COUNT(*) as total_courses, COUNT(CASE WHEN prefix != '' THEN 1 END) as courses_with_prefix FROM courses;"

echo ""

# Show distinct prefixes
echo "📝 Distinct prefixes in courses table:"
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME -e "SELECT DISTINCT prefix FROM courses WHERE prefix != '' ORDER BY prefix;"

echo ""
echo "✅ Verification complete!"