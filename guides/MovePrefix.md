# Database Migration: Move Prefix from Schedules to Courses

This migration moves the prefix column from the schedules table to the courses table to better align with the application's data model.

## Files Created

- `scripts/migrations/001_move_prefix_to_courses.sql` - Main migration SQL script
- `scripts/run_prefix_migration.sh` - Automated migration runner with backup
- `scripts/verify_prefix_migration.sh` - Verification script to check results

## What This Migration Does

### 1. **Adds prefix column to courses table**
```sql
ALTER TABLE courses ADD COLUMN prefix VARCHAR(10) NOT NULL DEFAULT '';
```

### 2. **Migrates existing data**
```sql
UPDATE courses c 
JOIN schedules s ON c.schedule_id = s.id 
JOIN prefixes p ON s.prefix_id = p.id 
SET c.prefix = p.prefix;
```

### 3. **Removes prefix_id from schedules table**
```sql
ALTER TABLE schedules DROP FOREIGN KEY schedules_ibfk_1;
ALTER TABLE schedules DROP COLUMN prefix_id;
```

### 4. **Adds performance index**
```sql
ALTER TABLE courses ADD INDEX idx_courses_prefix (prefix);
```

## How to Run the Migration

### Option 1: Automated Script (Recommended)
```bash
# Make scripts executable
chmod +x scripts/run_prefix_migration.sh
chmod +x scripts/verify_prefix_migration.sh

# Run the migration
./scripts/run_prefix_migration.sh
```

### Option 2: Manual SQL Execution
```bash
# Create backup first
mysqldump -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME > backup_before_migration.sql

# Run migration
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < scripts/migrations/001_move_prefix_to_courses.sql
```

## Verification

After running the migration, verify the results:

```bash
# Run verification script
./scripts/verify_prefix_migration.sh

# Or manual verification
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME
mysql> SELECT COUNT(*) FROM courses WHERE prefix != '';
mysql> DESCRIBE courses;
mysql> DESCRIBE schedules;
```

## Expected Results

- ✅ All courses should have prefix values populated
- ✅ Schedules table should no longer have prefix_id column
- ✅ Courses table should have new prefix column with index
- ✅ No data loss should occur

## Rollback Instructions

If you need to rollback the migration:

```bash
# Restore from backup
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD $DB_NAME < backups/before_prefix_migration_YYYYMMDD_HHMMSS.sql
```

## Safety Features

- **Automatic Backup**: Creates timestamped backup before migration
- **Validation**: Checks for required environment variables and files
- **Confirmation**: Prompts user before making changes
- **Error Handling**: Stops on any error and provides rollback instructions
- **Verification**: Provides tools to verify migration success

## After Migration

Once the migration is complete, you'll need to update your application code to:

1. ✅ Update Go structs to reflect new schema
2. ✅ Modify database queries to use courses.prefix instead of joins
3. ✅ Update forms and templates that reference schedule prefixes
4. ✅ Test all functionality to ensure proper operation

The migration scripts handle only the database changes. Application code updates must be done separately.