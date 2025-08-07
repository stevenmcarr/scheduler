# WMU Course Scheduler Unit Tests

This directory contains a comprehensive suite of unit tests for the WMU Course Scheduler application's controller functions. The tests are designed to ensure that changes to the codebase don't break existing functionality.

## 📁 Test Structure

```
unit_tests/
├── controllers_auth_test.go      # Authentication-related controller tests
├── controllers_data_test.go      # Data management controller tests  
├── controllers_schedule_test.go  # Schedule and conflict management tests
├── crosslisting_conflicts_test.go # Crosslisting conflict detection tests
├── course_conflicts_test.go      # Course conflict detection tests [NEW]
├── test_utils.go                 # Common testing utilities
├── run_tests.sh                  # Test runner script
└── README.md                     # This file
```

## 🧪 Test Categories

### 1. Authentication Controllers (`controllers_auth_test.go`)
Tests for user authentication, session management, and security features:

- **LoginController**: Tests login functionality with valid/invalid credentials
- **SignupController**: Tests user registration with validation
- **LogoutController**: Tests logout and session cleanup
- **SessionManagement**: Tests protected routes and session handling
- **FormValidation**: Tests email and username validation

**Key Test Cases:**
- Successful login with valid credentials
- Login failures with missing/invalid data
- User registration with duplicate detection
- Session-based access control
- Form data validation

### 2. Data Management Controllers (`controllers_data_test.go`)
Tests for CRUD operations on core data entities:

- **CoursesManagement**: Tests course creation, updating, and validation
- **RoomsManagement**: Tests room management with capacity validation
- **UserManagement**: Tests user updates including password changes
- **DataValidation**: Tests business logic validation for all entities

**Key Test Cases:**
- Saving courses with authentication
- Room addition with capacity validation
- User password changes with confirmation
- Bulk delete operations
- Data validation rules

### 3. Schedule & Conflict Management (`controllers_schedule_test.go`)
Tests for schedule operations and basic conflict detection:

- **ScheduleCreation**: Tests schedule generation and management
- **ConflictDetection**: Tests basic schedule conflict scenarios
- **CourseAssignment**: Tests course-to-schedule assignment
- **ValidationLogic**: Tests schedule validation rules

**Key Test Cases:**
- Schedule creation with course assignments
- Conflict detection between schedules
- Course validation during assignment
- Business rule enforcement

### 4. Crosslisting Conflicts (`crosslisting_conflicts_test.go`)
Tests for crosslisting conflict detection and resolution:

- **CrosslistingValidation**: Tests crosslisted course identification
- **ConflictResolution**: Tests conflict handling for crosslisted courses
- **EdgeCases**: Tests unusual crosslisting scenarios
- **ErrorHandling**: Tests error scenarios and recovery

**Key Test Cases:**
- Crosslisted course conflict detection
- Exception handling for crosslisted courses
- Complex crosslisting scenarios
- Error recovery and validation

### 5. Course Conflict Detection (`course_conflicts_test.go`) **[NEW]**
Comprehensive tests for the new course conflict detection system:

#### A. Course Number Extraction and Range Detection
- **`TestExtractNumericCourseNumber`**: Tests parsing of various course number formats
  - Basic numbers (2150), honors courses (2150H), writing intensive (2150W)
  - Edge cases with mixed alphanumeric strings and prefixes
- **`TestIsInSameCourseRange`**: Tests course range conflict logic
  - Range grouping (1000-1999, 2000-2999, 3000-3999, 5000-5999, 6000-6999)
  - Cross-range conflict prevention
- **`TestCourseNumberRangeBoundaries`**: Tests boundary conditions
  - Range edge values and boundary crossings
  - Gap handling (4000s range not covered)

#### B. Time Slot Overlap Detection
- **`TestTimeSlotsOverlap`**: Tests time slot overlap logic
  - Overlapping times on same/different days
  - Adjacent time slots (no false positives)
- **`TestTimeSlotEdgeCases`**: Tests complex time scenarios
  - Exact boundary conditions, empty time strings
  - Long time slots containing shorter ones

#### C. Prerequisite Chain Detection
- **`TestAreCoursesOnSamePrerequisiteChain`**: Tests prerequisite relationships
  - Direct prerequisite relationships
  - Multi-step prerequisite chains
- **`TestComplexPrerequisiteChains`**: Tests advanced scenarios
  - Long prerequisite chains (3+ steps)
  - Complex branching and multiple chains

#### D. Exception Handling
- **`TestDetectCourseConflicts_CrosslistedExecution`**: Tests crosslisting exceptions
- **`TestDetectCourseConflicts_PrerequisiteChainException`**: Tests prerequisite exceptions
- **`TestDetectCourseConflicts_EdgeCases`**: Tests unusual scenarios
  - Empty course lists, duplicate CRNs, nil time slots

#### E. Integration and Performance
- **`TestDetectCourseConflicts_NoExceptions`**: End-to-end conflict detection
- **`TestDetectCourseConflicts_MultipleConflicts`**: Multiple conflict scenarios
- **`BenchmarkDetectCourseConflicts`**: Performance testing with large datasets

## 🚀 Running Tests

### Run All Tests
```bash
cd unit_tests
./run_tests.sh
```

### Run Specific Test Categories

#### Course Conflict Tests (NEW)
```bash
go test -v ./unit_tests/ -run "TestExtractNumericCourseNumber|TestIsInSameCourseRange|TestTimeSlotsOverlap|TestAreCoursesOnSamePrerequisiteChain|TestDetectCourseConflicts" -timeout 30s
```

#### Authentication Tests
```bash
go test -v ./unit_tests/ -run "Test.*Auth" -timeout 30s
```

#### Data Management Tests
```bash
go test -v ./unit_tests/ -run "Test.*Data" -timeout 30s
```

#### Schedule Tests
```bash
go test -v ./unit_tests/ -run "Test.*Schedule" -timeout 30s
```

#### Business Logic Tests
```bash
go test -v ./unit_tests/ -run "Test.*Logic|Test.*Validation" -timeout 30s
```

#### Session Tests
```bash
go test -v ./unit_tests/ -run "Test.*Session" -timeout 30s
```

### Run Tests with Coverage
```bash
go test -v -cover ./unit_tests/ -timeout 60s -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run Performance Benchmarks
```bash
go test -bench=BenchmarkDetectCourseConflicts ./unit_tests/ -benchtime=5s
```

## 📊 Test Coverage

### Course Conflict Detection Coverage (NEW)
- ✅ **Course Number Parsing**: 9 test cases covering all formats
- ✅ **Range Detection**: 11 test cases including boundaries  
- ✅ **Time Overlap**: 6 test cases with edge scenarios
- ✅ **Prerequisite Chains**: 6 test cases with complex chains
- ✅ **Exception Handling**: 2 test cases for both exception types
- ✅ **Edge Cases**: 4 test cases for unusual scenarios
- ✅ **Integration**: 7 test cases for end-to-end functionality
- ✅ **Performance**: 1 benchmark test with large datasets

### Overall Test Metrics
- **Total Test Cases**: 75+ individual test cases
- **Test Categories**: 6 major functional areas
- **Coverage Areas**: Authentication, Data Management, Scheduling, Conflicts
- **Performance Tests**: Benchmark validation included
- **Edge Case Coverage**: Comprehensive boundary testing

### Traditional Coverage Areas
- **Authentication and Authorization**: Login, logout, session management, admin checks
- **Data Management**: CRUD operations for all major entities
- **Schedule Management**: Schedule creation, course assignment, validation
- **Conflict Detection**: Course conflicts, room conflicts, instructor conflicts, crosslisting conflicts
- **Business Logic**: Validation rules, data integrity, error handling
- **Session Management**: Session creation, expiration, security

## 🛠️ Mock Objects and Test Utilities

### MockCourseConflictScheduler (NEW)
- Implements course conflict detection interface
- Provides controllable mock responses for prerequisites and crosslistings
- Supports complex prerequisite chain simulation
- Enables isolated unit testing of conflict detection algorithms

### Test Router Setup
- Gin router with test mode configuration
- Session middleware with test store
- CSRF protection disabled for simplified testing

### Authentication Helpers
- `CreateAuthenticatedSession()`: Creates authenticated test sessions
- Session cookie management for stateful testing

### Test Data Creation
- Helper functions for creating test courses, instructors, rooms
- Mock data generation for consistent testing
- Course conflict test data structures

## 🔧 Integration with CI/CD

The test suite is designed for easy integration with continuous integration systems:

### GitHub Actions Example
```yaml
- name: Run Unit Tests
  run: |
    cd unit_tests
    ./run_tests.sh
    
- name: Generate Coverage Report
  run: |
    go test -cover ./unit_tests/ -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html

- name: Run Performance Benchmarks
  run: |
    go test -bench=BenchmarkDetectCourseConflicts ./unit_tests/ -benchtime=5s
```

### Test Results Format
- **PASS/FAIL** status for each test category
- **Coverage percentage** reporting
- **Performance benchmarks** with timing data
- **Detailed error reporting** for failed tests

## 🎯 Best Practices

### Writing New Tests
1. **Follow naming conventions**: `Test<FunctionName>_<Scenario>`
2. **Use table-driven tests** for multiple input scenarios
3. **Include edge cases** and boundary conditions
4. **Mock external dependencies** for isolation
5. **Add performance tests** for critical algorithms

### Mock Object Guidelines
1. **Setup appropriate expectations** with `On()` calls
2. **Use `mock.AnythingOfType()`** for flexible parameter matching
3. **Reset mocks** between test cases when needed
4. **Verify mock expectations** are met

### Test Maintenance
1. **Update tests** when functionality changes
2. **Add regression tests** for bug fixes
3. **Maintain test documentation** and examples
4. **Review test coverage** regularly

## 🚨 Recent Updates

### Course Conflict Detection Testing (NEW)
Added comprehensive unit tests for the new course conflict detection system:

- **46 individual test cases** covering all aspects of course conflict detection
- **Mock-based testing** for isolated unit testing without database dependencies
- **Performance benchmarking** to ensure scalability with large course lists
- **Edge case coverage** including boundary conditions and error scenarios
- **Integration testing** validating end-to-end conflict detection workflow

### Enhanced Test Runner
Updated `run_tests.sh` to include course conflict detection tests as a separate category with dedicated reporting.

The unit test suite now provides complete validation coverage for both existing functionality and the new course conflict detection system, ensuring robust operation of the WMU Course Scheduler application.
