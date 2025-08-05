# WMU Course Scheduler Unit Tests

This directory contains a comprehensive suite of unit tests for the WMU Course Scheduler application's controller functions. The tests are designed to ensure that changes to the codebase don't break existing functionality.

## 📁 Test Structure

```
unit_tests/
├── controllers_auth_test.go      # Authentication-related controller tests
├── controllers_data_test.go      # Data management controller tests  
├── controllers_schedule_test.go  # Schedule and conflict management tests
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
Tests for advanced scheduling functionality:

- **ScheduleManagement**: Tests schedule CRUD operations
- **ConflictDetection**: Tests time and resource conflict detection
- **CrosslistingManagement**: Tests course crosslisting functionality
- **APIEndpoints**: Tests API endpoints for schedule data
- **BusinessLogic**: Tests scheduling business rules

**Key Test Cases:**
- Schedule creation and deletion
- Conflict detection algorithms
- Crosslisting management with validation
- API response formatting
- Schedule filtering and querying

## 🚀 Running Tests

### Option 1: Use the Test Runner Script (Recommended)
```bash
cd /Users/carr/scheduler
./unit_tests/run_tests.sh
```

The script provides:
- ✅ Organized test execution by category
- 📊 Coverage reporting
- 🎨 Color-coded output
- 📋 Detailed summary

### Option 2: Run Individual Test Categories
```bash
# Authentication tests only
go test -v ./unit_tests/ -run "Test.*Auth.*"

# Data management tests only
go test -v ./unit_tests/ -run "Test.*Management|Test.*Data.*"

# Schedule and conflict tests only
go test -v ./unit_tests/ -run "Test.*Schedule|Test.*Conflict|Test.*Crosslisting"

# Business logic tests only
go test -v ./unit_tests/ -run "Test.*Validation|Test.*Logic"
```

### Option 3: Run All Tests
```bash
cd /Users/carr/scheduler
go test -v ./unit_tests/
```

### Option 4: Run with Coverage
```bash
cd /Users/carr/scheduler
go test -v -cover ./unit_tests/ -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## 📊 Test Coverage

The tests aim to cover:

- ✅ **Authentication flows**: Login, logout, signup
- ✅ **Data validation**: All input validation rules
- ✅ **CRUD operations**: Create, read, update, delete for all entities
- ✅ **Error handling**: Invalid inputs, missing data, unauthorized access
- ✅ **Session management**: Authentication state, session persistence
- ✅ **Business logic**: Scheduling rules, conflict detection
- ✅ **API endpoints**: JSON responses, parameter validation

## 🏗️ Test Architecture

### Mock Strategy
Tests use **lightweight mocking** with simplified handlers that mirror actual controller behavior without requiring a full database connection.

### Session Management
Tests include session middleware setup to properly test authentication-dependent functionality.

### Test Data
Each test file includes mock data structures that mirror the application's actual data models.

### Utilities
Common testing utilities are centralized in `test_utils.go` for:
- Test router setup with sessions
- Authenticated session creation
- Common test data structures

## 🔧 Adding New Tests

### For New Controller Functions:
1. Identify the appropriate test file based on functionality
2. Add test cases following the existing patterns
3. Include both positive and negative test cases
4. Test authentication requirements where applicable

### Test Function Naming Convention:
```go
func TestFunctionName(t *testing.T) {
    t.Run("Specific Test Case", func(t *testing.T) {
        // Test implementation
    })
}
```

### Example Test Structure:
```go
func TestNewController(t *testing.T) {
    router := SetupTestRouter()
    
    // Setup mock handler
    mockHandler := func(c *gin.Context) {
        // Mock implementation
    }
    router.POST("/endpoint", mockHandler)
    
    t.Run("Valid Input", func(t *testing.T) {
        cookie := CreateAuthenticatedSession(router, "testuser")
        // ... test implementation
        assert.Equal(t, http.StatusOK, w.Code)
    })
    
    t.Run("Invalid Input", func(t *testing.T) {
        // ... test implementation
        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}
```

## 🐛 Debugging Tests

### Common Issues:

1. **Session Issues**: Ensure `CreateAuthenticatedSession()` is called before testing protected routes
2. **JSON Binding**: Verify Content-Type headers are set correctly for JSON requests
3. **Form Data**: Use proper form encoding for POST requests
4. **Route Setup**: Ensure routes are registered before testing

### Debug Tips:
```bash
# Run specific test with verbose output
go test -v ./unit_tests/ -run "TestSpecificFunction"

# Run with additional debugging
go test -v ./unit_tests/ -run "TestSpecificFunction" -test.v

# Check test coverage for specific function
go test -cover ./unit_tests/ -run "TestSpecificFunction"
```

## 📈 Integration with CI/CD

These tests are designed to integrate with continuous integration systems:

```bash
# In CI/CD pipeline
cd /path/to/scheduler
./unit_tests/run_tests.sh
```

The script returns appropriate exit codes:
- `0`: All tests passed
- `1`: Some tests failed

## 🔒 Security Testing

The test suite includes security-focused tests:

- **Authentication bypass attempts**
- **Session hijacking prevention**
- **Input validation and sanitization**
- **Authorization checks**
- **CSRF protection** (with middleware setup)

## 📝 Test Maintenance

### Regular Updates Needed:
1. **New Features**: Add tests for new controller functions
2. **API Changes**: Update tests when controller signatures change
3. **Business Rules**: Update validation tests when rules change
4. **Security Updates**: Add tests for new security measures

### Best Practices:
- Keep tests simple and focused
- Use descriptive test names
- Test both success and failure cases
- Maintain test data consistency
- Update tests when refactoring controllers

## 🤝 Contributing

When adding new controller functions:

1. **Write tests first** (TDD approach recommended)
2. **Test all edge cases** and error conditions
3. **Include authentication tests** for protected endpoints
4. **Document test cases** in code comments
5. **Run full test suite** before committing changes

## 📞 Support

For questions about the test suite:

1. Review existing test patterns in the codebase
2. Check this README for common scenarios
3. Run tests with `-v` flag for detailed output
4. Use the test runner script for organized execution

---

**Last Updated**: December 2024  
**Test Framework**: Go testing + Testify + Gin Test Mode  
**Coverage Target**: >80% of controller functions
