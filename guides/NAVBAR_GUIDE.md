# Navigation Bar Implementation Guide

## Overview
A top navigation bar has been implemented for the WMU Course Scheduler application. The navbar appears on all pages except login and signup pages and includes role-based access controls.

## Files Created/Modified

### 1. navbar.html - Navigation Component
- **Location**: `/Users/carr/scheduler/src/templates/navbar.html`
- **Purpose**: Reusable navigation component with role-based visibility
- **Features**:
  - WMU brown color scheme (#8B4513)
  - Responsive design
  - Admin-only sections
  - User welcome message and logout

### 2. Updated Pages with Navbar
All pages now include `{{template "navbar" .}}` and proper styling:

- **home.html** - Main dashboard
- **courses.html** - Course management
- **instructors.html** - Instructor management  
- **rooms.html** - Room management
- **timeslots.html** - Time slot management
- **users.html** - User management (admin only)
- **departments.html** - Department management (admin only)
- **prefixes.html** - Prefix management (admin only)

### 3. Base Template
- **base.html** - Common template structure for consistency

## Navigation Structure

### Regular Users
- **Home** - Dashboard with schedules
- **Courses** - View and manage courses
- **Instructors** - View instructors
- **Rooms** - View rooms and facilities
- **Timeslots** - View available time slots

### Administrators (Additional Access)
- **Users** - Manage user accounts and permissions
- **Departments** - Manage academic departments
- **Prefixes** - Manage course prefixes

## Implementation Details

### Data Requirements
Each page template expects a context object with:
```go
type PageData struct {
    User struct {
        Username      string
        Administrator bool
        // other user fields
    }
    // page-specific data
}
```

### URL Routes
The navbar expects these routes to be implemented:
- `/` - Home
- `/courses` - Courses
- `/instructors` - Instructors
- `/rooms` - Rooms
- `/timeslots` - Time slots
- `/users` - Users (admin only)
- `/departments` - Departments (admin only)
- `/prefixes` - Prefixes (admin only)
- `/logout` - Logout

### Styling Features
- **Responsive Design**: Mobile-friendly navigation
- **Role-based Styling**: Admin buttons have special styling
- **Hover Effects**: Interactive button states
- **University Colors**: WMU brown theme
- **Status Indicators**: Online/offline user status

### Security Notes
- Admin-only sections use `{{if .User.Administrator}}` conditionals
- Server-side route protection should complement client-side hiding
- User authentication state is displayed in the navbar

## Usage in Go Templates

### Including the Navbar
```html
{{template "navbar" .}}
```

### Page Structure
```html
{{define "page-name"}}
<!DOCTYPE html>
<html>
<head>
    <title>Page Title - WMU Course Scheduler</title>
    <!-- styles -->
</head>
<body>
    {{template "navbar" .}}
    <div class="content">
        <!-- page content -->
    </div>
</body>
</html>
{{end}}
```

### Handler Data Structure
```go
func pageHandler(w http.ResponseWriter, r *http.Request) {
    data := struct {
        User struct {
            Username      string
            Administrator bool
        }
        // other page data
    }{
        User: getCurrentUser(r),
        // populate other data
    }
    
    tmpl.ExecuteTemplate(w, "page-name", data)
}
```

## Next Steps

1. **Implement Route Handlers**: Create Go handlers for each navbar route
2. **Add Authentication Middleware**: Protect admin routes
3. **Database Integration**: Connect pages to your MySQL database
4. **Session Management**: Implement user sessions for navbar user info
5. **Testing**: Test responsive design and role-based access

## Benefits

- **Consistent Navigation**: Uniform experience across all pages
- **Role-based Access**: Secure admin functionality
- **Professional Appearance**: Clean, university-themed design
- **Mobile Responsive**: Works on all device sizes
- **Easy Maintenance**: Single navbar component for all pages
