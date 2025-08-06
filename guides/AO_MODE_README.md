# AO Mode Addition

This update adds "AO" (Asynchronous Online) as a new mode option for courses in the WMU Course Scheduler application.

## What Was Added

### Database Changes
- **sql/add_ao_mode.sql**: Migration script to add AO to the mode enum in the courses table
- **scripts/add_ao_mode.sh**: Automated script to run the database migration with backup

### Application Code Changes
- **src/db.go**: Updated Course struct documentation to include AO mode
- **src/templates/courses.html**: Added AO option to mode dropdown for editing courses
- **src/templates/add_course.html**: Added AO option to mode dropdown for adding new courses
- **unit_tests/controllers_data_test.go**: Updated mode validation tests to include AO

## Valid Modes

The application now supports six course delivery modes:

1. **IP** - In Person
2. **FSO** - Fully Synchronous Online  
3. **PSO** - Partially Synchronous Online
4. **H** - Hybrid
5. **CLAS** - CLAS (existing)
6. **AO** - Asynchronous Online (new)

## How to Apply the Changes

### 1. Run Database Migration

```bash
# Make the script executable (if not already done)
chmod +x scripts/add_ao_mode.sh

# Run the migration
./scripts/add_ao_mode.sh
```

This will:
- Create a backup of your database
- Add AO to the mode enum in the courses table
- Verify the change was successful

### 2. Restart Application

After running the database migration, restart your application server to use the updated templates:

```bash
# Stop current application
# Start application again
go run ./src/main.go
```

## Verification

### Database Verification
```sql
-- Check that AO was added to the mode enum
SHOW COLUMNS FROM courses LIKE 'mode';

-- Should show: enum('IP','FSO','PSO','H','CLAS','AO')
```

### Application Verification
1. Navigate to the courses page
2. Click on any course to edit it
3. Verify that the mode dropdown includes "AO" option
4. Navigate to "Add Course" page
5. Verify that the mode dropdown includes "AO" option

### Test Verification
```bash
# Run unit tests to verify mode validation
go test -v ./unit_tests/...

# All tests should pass, including the Mode Validation test
```

## Rollback Instructions

If you need to remove the AO mode:

```bash
# Restore from backup
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < backups/before_add_ao_mode_YYYYMMDD_HHMMSS.sql
```

## Files Modified

- `src/db.go` - Updated documentation
- `src/templates/courses.html` - Added AO dropdown option
- `src/templates/add_course.html` - Added AO dropdown option  
- `unit_tests/controllers_data_test.go` - Updated validation tests
- `sql/add_ao_mode.sql` - New migration script
- `scripts/add_ao_mode.sh` - New migration runner script

## Safety Features

- **Automatic Backup**: The migration script creates a timestamped backup before making changes
- **Connection Testing**: Verifies database connectivity before running migration
- **Error Handling**: Stops on any error with clear rollback instructions
- **Validation**: Unit tests verify the new mode works correctly

The AO mode is now fully functional and can be selected when creating or editing courses.
