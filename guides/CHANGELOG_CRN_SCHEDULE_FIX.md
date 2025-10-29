# Fix for CRN/Schedule_ID Unique Constraint

## Date: October 28, 2025

## Problem
After updating the courses table unique constraint from `crn` alone to `(crn, schedule_id)` pair, database operations were failing to insert courses with the same CRN but different schedule_id. The issue was that existing database routines were still querying and updating based on CRN alone, without considering schedule_id.

## Root Cause
The `AddOrUpdateCourse` function and `GetCourseDetailsByCRN` function in `src/db.go` were not accounting for the composite unique key `(crn, schedule_id)`:

1. **AddOrUpdateCourse**: The UPDATE query used `WHERE crn = ?` which would update ANY course with that CRN across all schedules
2. **AddOrUpdateCourse**: The existence check used `SELECT crn FROM courses WHERE crn = ?` which would prevent insertion if the CRN existed in ANY schedule
3. **GetCourseDetailsByCRN**: Queried by CRN alone, which could return courses from wrong schedules

## Changes Made

### 1. src/db.go - AddOrUpdateCourse function (lines 1037-1039)
**Before:**
```sql
UPDATE courses SET ... WHERE crn = ?
```

**After:**
```sql
UPDATE courses SET ... WHERE crn = ? AND schedule_id = ?
```

### 2. src/db.go - AddOrUpdateCourse function (lines 1055-1065)
**Before:**
```go
// Check if the CRN exists but no update was needed
var existingCRN int
err = scheduler.database.QueryRow("SELECT crn FROM courses WHERE crn = ?", crn).Scan(&existingCRN)
if err == nil {
    // CRN exists but no update was needed
    return nil
}
```

**After:**
```go
// Check if the CRN exists for this schedule but no update was needed
var existingCRN int
err = scheduler.database.QueryRow("SELECT crn FROM courses WHERE crn = ? AND schedule_id = ?", crn, scheduleID).Scan(&existingCRN)
if err == nil {
    // CRN exists in this schedule but no update was needed
    return nil
}
```

### 3. src/db.go - GetCourseDetailsByCRN function (line 1776)
**Before:**
```go
func (scheduler *wmu_scheduler) GetCourseDetailsByCRN(crn int) (CourseDetail, error)
```

**After:**
```go
func (scheduler *wmu_scheduler) GetCourseDetailsByCRN(crn int, scheduleID int) (CourseDetail, error)
```

**Query Before:**
```sql
WHERE c.crn = ? AND c.status != 'Deleted'
```

**Query After:**
```sql
WHERE c.crn = ? AND c.schedule_id = ? AND c.status != 'Deleted'
```

### 4. src/controllers.go - Crosslisting display (lines 4863, 4871)
**Before:**
```go
course1, err := scheduler.GetCourseDetailsByCRN(cl.CRN1)
course2, err := scheduler.GetCourseDetailsByCRN(cl.CRN2)
```

**After:**
```go
course1, err := scheduler.GetCourseDetailsByCRN(cl.CRN1, cl.ScheduleID1)
course2, err := scheduler.GetCourseDetailsByCRN(cl.CRN2, cl.ScheduleID2)
```

## Impact
- **Courses can now be properly inserted** with the same CRN across different schedules (e.g., CRN 10001 can exist in both Spring 2025 and Fall 2025)
- **Updates are now schedule-specific** - updating a course in one schedule won't affect the same CRN in another schedule
- **Crosslisting queries are accurate** - they now fetch the correct course from the correct schedule

## Testing Recommendations
1. Test importing/adding courses with duplicate CRNs across different schedules
2. Test updating courses with the same CRN in different schedules
3. Test crosslisting functionality with duplicate CRNs
4. Test schedule copying with duplicate CRNs

## Related Files
- `src/db.go` - Database access layer
- `src/controllers.go` - HTTP request handlers
- `sql/alter_courses_unique_constraint.sql` - Migration script that created the composite unique key
