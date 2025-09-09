# Department Column Added to Users Table

## Summary
Successfully added a department_id column to the users table that references the departments table as a foreign key. This allows users to be associated with specific departments in the WMU Course Scheduler application.

## Database Changes

### 1. Migration Script Created
- **File**: `sql/add_department_to_users.sql`
- **Purpose**: Adds department_id column with foreign key constraint to departments table
- **Features**:
  - Nullable column (allows users without departments)
  - Foreign key constraint with CASCADE update and SET NULL on delete
  - Index for performance

### 2. Migration Execution Script
- **File**: `scripts/add_department_to_users.sh`
- **Purpose**: Automated script to run migration on both dev and production databases
- **Status**: ✅ Successfully executed on both `wmu_schedules_dev` and `wmu_schedules` databases

## Code Changes

### 1. Database Layer (`src/db.go`)

#### User Struct Updated
```go
type User struct {
    ID            int
    Username      string
    Email         string
    Password      string
    IsLoggedIn    bool
    Administrator bool
    DepartmentID  *int // NEW: Nullable foreign key to departments table
}
```

#### Functions Updated
- **AddUser()**: Now accepts `departmentID *int` parameter
- **GetAllUsers()**: Updated SELECT query to include department_id
- **GetUserByUsername()**: Updated SELECT query to include department_id  
- **GetUserByEmail()**: Updated SELECT query to include department_id
- **UpdateUserByID()**: Now accepts `departmentID *int` parameter and updates department_id in database

### 2. Controller Layer (`src/controllers.go`)

#### AddUserGin Function
- Added department_id form field parsing
- Added department validation and conversion to nullable int pointer
- Updated AddUser() call to include department parameter
- Added departments to template data for dropdown

#### RenderUsersPageGin Function
- Added GetAllDepartments() call to fetch departments for dropdown
- Added departments to template data

#### SaveUsersGin Function
- Added department_id parsing from JSON user data
- Updated UpdateUserByID() call to include department parameter

#### renderAddUserFormGin Function
- Added GetAllDepartments() call
- Added departments to template data for form rendering

### 3. Template Updates

#### Add User Template (`src/templates/add_user.html`)
- Added department dropdown field after email field
- Department selection is optional (includes "Select a Department" option)
- Updated password requirements text to reflect new 8-character policy

#### Users Management Template (`src/templates/users.html`)  
- Added "Department" column header
- Added department dropdown for each user row
- Dropdown shows current user's department as selected
- Uses template logic to properly handle nullable department IDs

## Features

### 1. Department Assignment
- **Add User Page**: Administrators can assign departments when creating new users
- **Users Management Page**: Administrators can update department assignments for existing users
- **Optional Assignment**: Users are not required to have a department (nullable field)

### 2. Data Integrity
- **Foreign Key Constraint**: Ensures department_id references valid department
- **Cascade Behavior**: 
  - Updates cascade when department ID changes
  - Sets to NULL when department is deleted (preserves user data)

### 3. User Experience
- **Dropdown Selection**: Easy department selection via dropdown menus
- **Visual Feedback**: Current departments are properly displayed and selected
- **Form Validation**: Maintains existing user validation while adding department support

## Database Schema

### Users Table Structure (After Migration)
```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE, 
    password VARCHAR(255) NOT NULL,
    is_logged_in BOOLEAN DEFAULT FALSE,
    administrator BOOLEAN DEFAULT FALSE,
    department_id INT NULL,
    FOREIGN KEY (department_id) REFERENCES departments(id) 
        ON DELETE SET NULL ON UPDATE CASCADE,
    INDEX idx_users_department_id (department_id)
);
```

## Testing Status

### ✅ Completed
- Database migration executed successfully
- Code compilation successful
- All function signatures updated consistently
- Template updates completed

### 🔄 Next Steps
1. Test user creation with department assignment
2. Test user editing with department changes  
3. Verify foreign key constraints work correctly
4. Test department deletion behavior (should set user department_id to NULL)
5. Consider adding department filtering/searching capabilities

## Files Modified
- `src/db.go` - User struct and database functions
- `src/controllers.go` - User management controllers
- `src/templates/add_user.html` - Add user form
- `src/templates/users.html` - Users management page
- `sql/add_department_to_users.sql` - Database migration
- `scripts/add_department_to_users.sh` - Migration execution script

## Backward Compatibility
- Existing users will have NULL department_id (no department assigned)
- All existing functionality remains intact
- No breaking changes to API or user experience
