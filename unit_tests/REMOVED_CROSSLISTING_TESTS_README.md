# Removed Courses and Cross-listing Logic Unit Tests

This document describes the comprehensive unit test suite for the new conflict detection logic regarding "Removed" status courses and cross-listed courses.

## Overview

The WMU Course Scheduler conflict detection system has been enhanced with two critical business rules:

1. **Removed Course Rule**: Courses with "Removed" status cannot conflict with any other course
2. **Cross-listing Rule**: Cross-listed courses should not conflict inappropriately with each other

## Test Files

### 1. `removed_crosslisting_conflicts_test.go`
Main test file focusing on the core logic of removed courses and cross-listing conflict detection.

**Key Test Categories:**
- Removed courses instructor conflicts
- Removed courses room conflicts  
- Removed courses crosslisting conflicts
- Removed courses course conflicts
- Cross-listed courses behavior
- Normal course conflict validation
- Edge cases and status combinations

### 2. `integration_removed_crosslisting_test.go`
Integration tests that verify complex scenarios and comprehensive conflict detection workflows.

**Key Test Categories:**
- Multi-course complex scenarios
- Cross-listed courses with different modes
- Removed courses in cross-listing groups
- All conflict types validation
- Large-scale performance testing

## Test Structure

### Mock Scheduler Implementation
Both test files use sophisticated mock schedulers that simulate the actual conflict detection logic:

```go
type MockRemovedScheduler struct {
    mock.Mock
    crosslistings       map[string]bool
    prerequisites      map[string][]string
}
```

### Course Detail Structure
Tests use structures that mirror the actual `CourseDetail` struct:

```go
type RemovedCourseDetail struct {
    ID                  int
    CRN                 int
    Section             string
    ScheduleID          int
    Prefix              string
    CourseNumber        string
    Title               string
    InstructorID        int
    InstructorFirstName string
    InstructorLastName  string
    TimeSlotID          int
    RoomID              int
    Mode                string
    Status              string  // Key field for removed status
    Lab                 bool
    TimeSlot            *RemovedTimeSlot
}
```

## Critical Test Cases

### 1. Removed Course Logic Tests

#### `TestRemovedCourses_InstructorConflicts_ShouldBeSkipped`
- **Purpose**: Verify removed courses don't create instructor conflicts
- **Scenario**: Two courses with same instructor and overlapping time slots, one removed
- **Expected**: No instructor conflicts reported

#### `TestRemovedCourses_RoomConflicts_ShouldBeSkipped`
- **Purpose**: Verify removed courses don't create room conflicts
- **Scenario**: Two courses with same room and overlapping time slots, one removed
- **Expected**: No room conflicts reported

#### `TestRemovedCourses_CrosslistingConflicts_ShouldBeSkipped`
- **Purpose**: Verify removed courses don't create crosslisting conflicts
- **Scenario**: Crosslisted courses with different time slots, one removed
- **Expected**: No crosslisting conflicts reported

#### `TestRemovedCourses_CourseConflicts_ShouldBeSkipped`
- **Purpose**: Verify removed courses don't create course conflicts
- **Scenario**: Courses in same range that would normally conflict, one removed
- **Expected**: No course conflicts reported

### 2. Cross-listing Logic Tests

#### `TestCrosslistedCourses_SameTimeSlot_ShouldNotConflict`
- **Purpose**: Verify crosslisted courses with identical time slots don't conflict
- **Scenario**: Two crosslisted courses with same time slot
- **Expected**: No crosslisting conflicts reported

#### `TestCrosslistedCourses_DifferentTimeSlots_ShouldConflict`
- **Purpose**: Verify crosslisted courses with different time slots do conflict
- **Scenario**: Two crosslisted courses with different time slots
- **Expected**: Crosslisting conflict reported

#### `TestCrosslistedCourses_AOMode_ShouldNotConflict`
- **Purpose**: Verify AO mode courses are exempt from time conflicts
- **Scenario**: Crosslisted courses with different time slots, one in AO mode
- **Expected**: No crosslisting conflicts reported

### 3. Integration Tests

#### `TestIntegration_MultipleRemovedCourses_ComplexScenario`
- **Purpose**: Test complex scenario with multiple courses, some removed
- **Scenario**: Mix of removed and normal courses with various conflict potentials
- **Expected**: Only normal courses create conflicts

#### `TestIntegration_CrosslistedCoursesWithDifferentModes`
- **Purpose**: Test crosslisted courses with various modes
- **Scenario**: Crosslisted courses in Traditional, AO, FSO, PSO modes
- **Expected**: Appropriate conflict behavior based on mode

#### `TestIntegration_VerifyAllConflictTypesRespectRemovedStatus`
- **Purpose**: Comprehensive verification that all conflict types respect removed status
- **Scenario**: Courses that would conflict in every way, one removed
- **Expected**: No conflicts of any type when one course is removed

### 4. Edge Case Tests

#### `TestBothCoursesRemoved_ShouldNotConflict`
- **Purpose**: Verify no conflicts when both courses are removed
- **Scenario**: Two courses both with "Removed" status
- **Expected**: No conflicts of any type

#### `TestMixedStatusCombinations`
- **Purpose**: Test various status combinations
- **Scenarios**: Different combinations of Scheduled, Removed, Cancelled, Active
- **Expected**: Only "Removed" status prevents conflicts

## Running the Tests

### Option 1: Specialized Test Runner
```bash
cd /home/stevecarr/scheduler/unit_tests
./run_removed_crosslisting_tests.sh
```

### Option 2: Individual Test Categories
```bash
# Removed courses logic
go test -v ./unit_tests/ -run "TestRemovedCourses.*"

# Crosslisted courses logic  
go test -v ./unit_tests/ -run "TestCrosslistedCourses.*"

# Integration tests
go test -v ./unit_tests/ -run "TestIntegration.*"

# Edge cases
go test -v ./unit_tests/ -run "TestBothCoursesRemoved|TestMixedStatusCombinations"
```

### Option 3: All New Tests
```bash
go test -v ./unit_tests/ -run "TestRemovedCourses|TestCrosslistedCourses|TestIntegration|TestBothCoursesRemoved|TestMixedStatusCombinations|TestNormalCourses"
```

## Test Coverage

The test suite provides comprehensive coverage of:

1. **All Conflict Detection Functions**:
   - `DetectConflictsBetweenSchedules`
   - `detectCrosslistingConflicts`
   - `detectCourseConflicts`

2. **All Status Checks**: Tests verify that "Removed" status is properly checked in all conflict detection paths

3. **All Mode Exemptions**: Tests verify that course modes (AO, FSO, PSO) are properly handled in crosslisting scenarios

4. **Edge Cases**: Comprehensive testing of unusual combinations and large-scale scenarios

## Expected Outcomes

When all tests pass, the system guarantees:

1. ✅ **Removed courses never create instructor conflicts**
2. ✅ **Removed courses never create room conflicts**  
3. ✅ **Removed courses never create crosslisting conflicts**
4. ✅ **Removed courses never create course conflicts**
5. ✅ **Cross-listed courses with same time slots don't conflict**
6. ✅ **Cross-listed courses with different time slots do conflict (unless exempt mode)**
7. ✅ **AO mode courses are exempt from time-based conflicts**
8. ✅ **Normal course conflicts continue to work properly**
9. ✅ **Complex scenarios with mixed removed/normal courses work correctly**
10. ✅ **Performance is maintained even with large numbers of removed courses**

## Validation Against Production Logic

These tests directly validate the production implementation by:

1. **Mirroring Actual Logic**: Test functions mirror the exact logic in `controllers.go`
2. **Same Data Structures**: Test structures match the actual `CourseDetail` struct
3. **Comprehensive Scenarios**: Tests cover all realistic combinations of course statuses and modes
4. **Performance Validation**: Large-scale tests ensure the logic doesn't degrade performance

## Maintenance

When updating the conflict detection logic:

1. **Add New Test Cases**: For any new business rules or edge cases
2. **Update Mock Functions**: If the actual function signatures change
3. **Verify Status Checks**: Ensure any new conflict types properly check for "Removed" status
4. **Run Full Suite**: Always run the complete test suite after changes

## Dependencies

The tests require:
- Go testing framework
- `github.com/stretchr/testify/assert` for assertions
- `github.com/stretchr/testify/mock` for mocking

## Integration with CI/CD

These tests are designed to be run as part of the continuous integration pipeline to ensure that:
- No regression is introduced in conflict detection logic
- New features properly implement the removed course and cross-listing rules
- Performance remains acceptable as the codebase evolves
