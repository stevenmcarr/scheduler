# Administrator Restrictions for Prerequisites System

## Overview

The Prerequisites Management system has been configured to restrict access to administrators only, ensuring that only authorized users can modify course prerequisite relationships.

## Security Implementation

### Authentication Checks
All prerequisite controller functions now include:

1. **User Authentication**: Verifies the user is logged in
2. **Administrator Verification**: Confirms the user has administrator privileges
3. **Access Denial**: Returns 403 Forbidden for non-administrators

### Protected Functions

The following functions are now administrator-only:

- `RenderPrerequisitesPageGin()` - View prerequisites page
- `FilterPrerequisitesGin()` - Filter prerequisites by course number
- `AddPrerequisiteGin()` - Add new prerequisites
- `UpdatePrerequisiteGin()` - Modify existing prerequisites
- `DeletePrerequisiteGin()` - Delete prerequisites

### Navigation Changes

**Before**: Prerequisites link was in the main navigation (visible to all users)
**After**: Prerequisites link is in the Administrator section (visible only to administrators)

```html
<!-- Admin section in navbar -->
{{if and .User .User.Administrator}}
<div class="admin-section">
    <a href="/scheduler/departments" class="navbar-item admin-item">Departments</a>
    <a href="/scheduler/prefixes" class="navbar-item admin-item">Prefixes</a>
    <a href="/scheduler/prerequisites" class="navbar-item admin-item">Prerequisites</a>
    <a href="/scheduler/users" class="navbar-item admin-item">Users</a>
</div>
{{end}}
```

## Error Handling

Non-administrator users attempting to access prerequisite functions will receive:

- **HTTP Status**: 403 Forbidden
- **Error Message**: "Access denied. Administrator privileges required."
- **Redirect**: Users are redirected to login if not authenticated

## Files Modified

### Controller Functions (`src/controllers.go`)
Added administrator checks to all prerequisite functions:
- Authentication verification using `getCurrentUser()`
- Administrator privilege checking
- Consistent error handling and messaging

### Navigation Template (`src/templates/navbar.html`)
- Moved Prerequisites link from general navigation to admin section
- Link only visible to authenticated administrators

### Documentation Updates
- Updated README files to reflect administrator-only access
- Added security implementation details

## Testing Administrator Access

### As Administrator
1. Log in with administrator credentials
2. Prerequisites link appears in admin section of navbar
3. Full access to prerequisites functionality

### As Regular User
1. Log in with non-administrator credentials
2. Prerequisites link is not visible in navbar
3. Direct URL access returns "Access denied" error

### Not Logged In
1. Attempting to access prerequisites URLs redirects to login page
2. Must authenticate before accessing any prerequisite functions

## Benefits

1. **Data Integrity**: Only authorized personnel can modify prerequisite relationships
2. **Security**: Prevents unauthorized changes to academic requirements
3. **Audit Trail**: Clear separation between administrator and user capabilities
4. **Consistency**: Follows same security pattern as other admin functions (departments, prefixes, users)

## API Endpoint Security

All prerequisite endpoints now require administrator authentication:

- `GET /scheduler/prerequisites` - Administrator only
- `POST /scheduler/prerequisites` - Administrator only (filtering)
- `POST /scheduler/add_prerequisite` - Administrator only
- `POST /scheduler/update_prerequisite` - Administrator only
- `POST /scheduler/delete_prerequisite` - Administrator only

Direct API access without proper authentication returns appropriate HTTP status codes and error messages.
