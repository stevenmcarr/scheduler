# Prerequisites Management System

This document describes the Prerequisites Management system that has been added to the WMU Course Scheduler.

## Overview

The Prerequisites system allows administrators to manage course prerequisite relationships, showing which courses are required before taking other courses. The system provides a web interface with filtering capabilities and editable tables.

**Note: Prerequisites management is restricted to administrators only.**

## Features

### 📋 Prerequisites Table
- **Predecessor Course**: The course that must be taken first (prerequisite)
- **Successor Course**: The course that requires the prerequisite
- **Editable Fields**: All fields can be edited inline
- **Course Number Filter**: Filter prerequisites by course number
- **Dropdown Menus**: Prefix fields use dropdown menus populated from the database

### 🔍 Filtering
- Filter by course number to find all prerequisites containing that number
- Search works for both predecessor and successor course numbers
- Clear filter option to view all prerequisites

### ✏️ Editing
- **Inline Editing**: Click "Edit" to modify prerequisites directly in the table
- **Dropdown Prefixes**: Course prefixes are selected from dropdown menus
- **Number Fields**: Course numbers are text fields for flexibility
- **Save/Cancel**: Save changes or cancel to revert to original values
- **Delete**: Remove prerequisites with confirmation dialog

### ➕ Adding New Prerequisites
- Add new prerequisite relationships using the form at the top
- All fields are required (predecessor prefix/number, successor prefix/number)
- Dropdown menus for prefixes ensure consistency

## Database Schema

### Prerequisites Table
```sql
CREATE TABLE prerequisites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    predecessor_prefix VARCHAR(10) NOT NULL,
    predecessor_number VARCHAR(10) NOT NULL,
    successor_prefix VARCHAR(10) NOT NULL,
    successor_number VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_predecessor (predecessor_prefix, predecessor_number),
    INDEX idx_successor (successor_prefix, successor_number),
    UNIQUE KEY unique_prerequisite (predecessor_prefix, predecessor_number, successor_prefix, successor_number)
);
```

### Prefixes Table
```sql
CREATE TABLE prefixes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    prefix VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Setup Instructions

### 1. Database Setup
Run the database setup script:
```bash
./setup_prerequisites.sh
```

Or manually execute the SQL:
```bash
mysql -u [username] -p [database_name] < create_prerequisites_table.sql
```

### 2. Build and Run
```bash
go build -o bin/course_scheduler src/*.go
./bin/course_scheduler
```

### 3. Access Prerequisites
Navigate to: `http://localhost:8080/scheduler/prerequisites`

Or use the "Prerequisites" link in the **Administrator section** of the navigation bar.

**Important**: You must be logged in as an administrator to access the Prerequisites functionality. Non-administrator users will receive an "Access denied" error.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/scheduler/prerequisites` | View all prerequisites |
| POST | `/scheduler/prerequisites` | Filter prerequisites by course number |
| POST | `/scheduler/add_prerequisite` | Add new prerequisite |
| POST | `/scheduler/update_prerequisite` | Update existing prerequisite |
| POST | `/scheduler/delete_prerequisite` | Delete prerequisite |

## Sample Data

The system includes sample prerequisites for testing:
- MATH 101 → MATH 102
- MATH 102 → MATH 201
- MATH 201 → MATH 301
- ENG 101 → ENG 201
- CS 101 → CS 201
- CS 201 → CS 301
- MATH 101 → CS 201 (cross-department prerequisite)

## Files Added/Modified

### New Files
- `src/templates/prereqs.html` - Prerequisites management page
- `create_prerequisites_table.sql` - Database schema and sample data
- `setup_prerequisites.sh` - Database setup script

### Modified Files
- `src/db.go` - Added Prerequisite struct and database functions
- `src/controllers.go` - Added prerequisites controller functions
- `src/routes.go` - Added prerequisites routes
- `src/templates/navbar.html` - Added Prerequisites navigation link

## Usage Examples

### Adding a Prerequisite
1. Select predecessor prefix from dropdown (e.g., "MATH")
2. Enter predecessor number (e.g., "101")
3. Select successor prefix from dropdown (e.g., "CS")
4. Enter successor number (e.g., "201")
5. Click "Add Prerequisite"

### Filtering Prerequisites
1. Enter course number in filter field (e.g., "101")
2. Click "Filter" to show all prerequisites containing "101"
3. Click "Clear Filter" to show all prerequisites

### Editing Prerequisites
1. Click "Edit" button for the prerequisite you want to modify
2. Use dropdown menus to change prefixes
3. Edit course numbers in text fields
4. Click "Save" to confirm changes or "Cancel" to revert

### Deleting Prerequisites
1. Click "Delete" button for the prerequisite you want to remove
2. Confirm deletion in the dialog box

## Technical Notes

- **CSRF Protection**: All forms include CSRF tokens for security
- **Responsive Design**: Interface adapts to mobile devices
- **Error Handling**: Graceful handling of database errors and validation
- **Unique Constraints**: Database prevents duplicate prerequisite relationships
- **Fallback Data**: System provides default prefixes if database tables don't exist

## Troubleshooting

### "No prerequisites found"
- Check if the prerequisites table exists and has data
- Run the setup script to create sample data

### "Failed to load prefixes"
- Ensure the prefixes table exists or that courses table has prefix data
- System will fall back to common prefixes if database query fails

### Database Connection Issues
- Verify database credentials in environment variables
- Check that MySQL service is running
- Ensure database exists and user has proper permissions

## Future Enhancements

Potential improvements for the prerequisites system:
- Import/export functionality for bulk prerequisite management
- Prerequisite validation when adding courses to schedules
- Visual prerequisite tree/graph display
- Integration with course catalog systems
- Prerequisite audit reports
