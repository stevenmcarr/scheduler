# Database Change: CRN Unique Constraint Update

## Date: October 16, 2025

## Change Summary
Modified the `courses` table unique constraint to allow CRN reuse across different schedules while maintaining uniqueness within each schedule.

## Previous Constraint
```sql
UNIQUE KEY `crn` (`crn`)
```
- CRN was globally unique across all schedules
- Same CRN could not be used in different terms/years
- Required offset when copying schedules (added 10000 to CRNs)

## New Constraint
```sql
UNIQUE KEY `unique_schedule_crn` (`schedule_id`, `crn`)
```
- CRN is now unique per schedule (composite key)
- Same CRN can be reused across different terms/years
- No offset needed when copying schedules

## Benefits
1. **Natural CRN Management** - CRNs can remain consistent across terms
2. **Simplified Schedule Copying** - No need to generate new CRNs
3. **Real-World Alignment** - Matches actual course scheduling practices where CRNs are reused each term
4. **Data Integrity** - Still prevents duplicate CRNs within a single schedule

## Files Modified

### 1. Database Migration Script
**File:** `sql/alter_courses_unique_constraint.sql`
- Drops old `crn` unique constraint
- Adds new `unique_schedule_crn (schedule_id, crn)` constraint

### 2. Database Schema Creation Script  
**File:** `sql/create_wmu_schedules_database.sql`
- Updated to include composite unique key in fresh installations

### 3. Copy Schedule Function
**File:** `src/db.go` - Function: `CopySchedule()`
- Removed CRN offset logic (was adding 10000)
- Now copies courses with original CRNs
- Comments updated to reflect new behavior

## Migration Applied
```bash
mysql -u wmu_cs -p wmu_schedules < sql/alter_courses_unique_constraint.sql
```

## Verification
```sql
-- Check the new constraint
SHOW INDEXES FROM courses WHERE Key_name = 'unique_schedule_crn';

-- Result shows composite key on (schedule_id, crn)
```

## Impact Analysis

### Positive Impacts
- ✅ Schedule copying now preserves original CRNs
- ✅ Same course can have same CRN across Fall 2025, Spring 2026, etc.
- ✅ More intuitive for users familiar with banner systems
- ✅ Reduced complexity in copy logic

### No Breaking Changes
- ✅ Existing data remains valid
- ✅ All existing courses maintain their CRNs
- ✅ Application logic updated accordingly
- ✅ No API changes required

## Testing Recommendations
1. Test schedule copying with identical CRNs
2. Verify CRN validation within same schedule
3. Confirm crosslisting functionality still works
4. Test conflict detection across schedules
5. Verify import functionality handles CRN correctly

## Rollback Plan
If needed, rollback with:
```sql
ALTER TABLE courses DROP INDEX unique_schedule_crn;
ALTER TABLE courses ADD UNIQUE KEY crn (crn);
```
Note: This requires ensuring no duplicate CRNs exist across schedules first.

## Related Features
- Schedule Copy functionality (uses this constraint)
- Course Import (validates CRNs per schedule)
- Conflict Detection (checks across schedules)
- Crosslisting (references CRNs)
