# Add User Utility

This utility allows you to add new users to the WMU Scheduler database with the same password encryption scheme as the main application, including the ability to grant administrator privileges.

## Features

- **Same Encryption**: Uses bcrypt with default cost for password hashing (identical to main application)
- **Email Validation**: Validates email format using regex pattern
- **Strong Password Requirements**: Enforces the same password policy as the application:
  - At least 15 characters long
  - Contains at least one uppercase letter
  - Contains at least one lowercase letter
  - Contains at least one number
  - Contains at least one special character
- **Administrator Privileges**: Option to create users with administrator access
- **Duplicate Prevention**: Checks for existing usernames and emails
- **Secure Input**: Passwords are entered securely (not echoed to terminal)
- **Environment Integration**: Reads database configuration from `.env` file

## Files

- `add_user.go` - Main Go program for adding users
- `add_user.sh` - Shell script wrapper for easy execution

## Prerequisites

- Go installed on your system
- Access to the MySQL database specified in the `.env` file
- The `.env` file must be present in the scheduler directory

## Usage

### Option 1: Using the Shell Script (Recommended)

```bash
./add_user.sh
```

The shell script will:
- Check for Go installation
- Verify `.env` file exists
- Check and install Go dependencies
- Run the add user program

### Option 2: Direct Go Execution

```bash
go run add_user.go
```

Or compile and run:
```bash
go build add_user.go
./add_user
```

## Interactive Prompts

When you run the script, you'll be prompted for:

1. **Username**: The login username for the new user
2. **Email**: Valid email address for the user
3. **Password**: Secure password (will not be displayed)
4. **Confirm Password**: Re-enter password for verification
5. **Administrator**: Choose whether to grant admin privileges (y/N)

## Example Session

```
WMU Scheduler - Add User Script
==============================
Connected to database: wmu_schedules_dev

Enter username: john_doe
Enter email: john.doe@wmich.edu
Enter password: [hidden input]
Confirm password: [hidden input]
Make user administrator? (y/N): y

User 'john_doe' added successfully as administrator!
```

## Administrator Privileges

Users created with administrator privileges can:
- Access administrative functions in the web application
- Manage other users
- Configure system settings
- Access all administrative pages

Regular users (non-administrators) have standard access to schedule viewing and editing functions.

## Error Handling

The script provides clear error messages for:
- **Database Connection Issues**: Problems connecting to MySQL
- **Invalid Email**: Email format validation failures
- **Weak Passwords**: Passwords not meeting security requirements
- **Password Mismatches**: Confirmation password doesn't match
- **Duplicate Users**: Username or email already exists
- **Database Errors**: Issues during user insertion

## Security Notes

- **Password Hashing**: Uses bcrypt with the same cost as the main application
- **No Plain Text Storage**: Passwords are never stored or logged in plain text
- **Environment Variables**: Database credentials read from `.env` file
- **Input Validation**: All inputs validated before database operations
- **Secure Entry**: Passwords entered without terminal echo

## Database Schema

The script inserts users into the `users` table with the following fields:
- `username` - Unique username for login
- `email` - User's email address
- `password` - bcrypt hashed password
- `administrator` - Boolean flag for admin privileges

## Troubleshooting

### Go Not Found
```bash
Error: Go is not installed or not in PATH
```
Install Go from https://golang.org/dl/

### Missing .env File
```bash
Error: .env file not found
```
Ensure the `.env` file exists in the scheduler directory with proper database credentials.

### Database Connection Failed
```bash
Failed to connect to database: dial tcp [::1]:3306: connect: connection refused
```
- Verify MySQL is running
- Check database credentials in `.env`
- Ensure database exists

### Dependency Issues
```bash
Error: Failed to resolve Go dependencies
```
Run `go mod tidy` to update dependencies, or use the shell script which handles this automatically.
