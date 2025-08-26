# Add User Script

This script allows you to add a new user to the WMU Scheduler database using the same password encryption scheme as the main application.

## Features

- Uses bcrypt with default cost for password hashing (same as the main application)
- Validates email format using regex
- Enforces strong password requirements:
  - At least 15 characters long
  - Contains at least one uppercase letter
  - Contains at least one lowercase letter
  - Contains at least one number
  - Contains at least one special character
- Checks for duplicate usernames and emails
- Secure password input (passwords are not echoed to terminal)
- Reads database configuration from the parent directory's `.env` file

## Prerequisites

- Go installed on your system
- Access to the MySQL database specified in the `.env` file
- The `.env` file must be present in the parent directory (`../`)

## Usage

1. Navigate to the scripts directory:
   ```bash
   cd /home/stevecarr/scheduler/scripts
   ```

2. Run the script:
   ```bash
   go run add_user.go
   ```
   
   Or compile and run:
   ```bash
   go build add_user.go
   ./add_user
   ```

3. Follow the prompts:
   - Enter username
   - Enter email address
   - Enter password (will not be displayed)
   - Confirm password (will not be displayed)

## Example

```
WMU Scheduler - Add User Script
==============================
Connected to database: wmu_schedules_dev

Enter username: admin
Enter email: admin@example.com
Enter password: [hidden input]
Confirm password: [hidden input]

User 'admin' added successfully!
```

## Error Handling

The script will display helpful error messages for:
- Database connection failures
- Invalid email addresses
- Weak passwords
- Password confirmation mismatches
- Duplicate usernames or emails
- Database insertion errors

## Security Notes

- Passwords are hashed using bcrypt with the same cost as the main application
- Passwords are not logged or displayed in plain text
- Database credentials are read from environment variables
- The script validates all input before attempting database operations
