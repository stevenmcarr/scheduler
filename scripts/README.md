# Development Database Scripts

This directory contains scripts to create and manage development databases for the WMU Course Scheduler application.

## Scripts

### 1. `create_dev_database.sh` - Full-featured database copy script

A comprehensive script that creates a development database by copying the production `wmu_schedules` database.

**Features:**
- Creates a complete copy of the production database
- Supports custom database names
- Can clean sensitive data for development use
- Creates backups during the process
- Generates development-specific .env file
- Comprehensive error handling and logging

**Usage:**
```bash
# Basic usage
./scripts/create_dev_database.sh -u dbuser -p password

# With all options
./scripts/create_dev_database.sh \
  --user dbuser \
  --password mypass \
  --host localhost \
  --port 3306 \
  --source-db wmu_schedules \
  --dev-db wmu_schedules_dev \
  --clean-sensitive \
  --force

# Get help
./scripts/create_dev_database.sh --help
```

**Options:**
- `-u, --user USER`: MySQL username (required)
- `-p, --password PASS`: MySQL password (will prompt if not provided)
- `-H, --host HOST`: MySQL host (default: localhost)
- `-P, --port PORT`: MySQL port (default: 3306)
- `-s, --source-db DB`: Source database name (default: wmu_schedules)
- `-d, --dev-db DB`: Development database name (default: wmu_schedules_dev)
- `-c, --clean-sensitive`: Clean sensitive data from development database
- `-f, --force`: Drop existing development database if it exists
- `--backup-file FILE`: Custom backup file path

### 2. `setup_dev_db.sh` - Simple wrapper script

A convenient wrapper that reads database credentials from your existing `.env` file.

**Usage:**
```bash
# Basic usage (reads credentials from .env)
./scripts/setup_dev_db.sh

# With additional options
./scripts/setup_dev_db.sh -c -f
./scripts/setup_dev_db.sh --clean-sensitive --force
./scripts/setup_dev_db.sh -d my_custom_dev_db

# Get help
./scripts/setup_dev_db.sh --help
```

**Requirements:**
- Must have a `.env` file in the project root with `DB_USER` and `DB_PASSWORD`
- Optional: `DB_HOST`, `DB_PORT`, `DB_NAME` (will use defaults if not set)

## What the Scripts Do

1. **Database Copy**: Creates an exact copy of your production database structure and data
2. **Development Environment**: Creates a `.env.dev` file configured for development
3. **Sensitive Data Cleaning** (if `-c` flag used):
   - Resets all user passwords to "password"
   - Changes email addresses to dev-safe format
   - Adds `[DEV]` prefix to schedule names
   - Creates a default admin user: `dev_admin` / `password`
4. **Backup Creation**: Creates a backup file before making changes
5. **Verification**: Verifies the copy was successful

## Example Workflows

### Quick Setup for Development
```bash
# Use the simple wrapper (easiest)
./scripts/setup_dev_db.sh -c -f

# This will:
# - Read your credentials from .env
# - Create wmu_schedules_dev database
# - Clean sensitive data
# - Create .env.dev file
```

### Custom Development Database
```bash
# Create a custom dev database with specific name
./scripts/create_dev_database.sh \
  -u myuser -p mypass \
  -d my_feature_dev \
  --clean-sensitive

# Use the development database
ENV_FILE=.env.dev ./bin/scheduler
```

### Testing Different Database Versions
```bash
# Create multiple dev databases for testing
./scripts/create_dev_database.sh -u dbuser -p pass -d test_v1 -c
./scripts/create_dev_database.sh -u dbuser -p pass -d test_v2 -c

# Modify .env.dev to point to different databases as needed
```

## Using the Development Database

After running the scripts, you'll have:

1. **Development Database**: `wmu_schedules_dev` (or custom name)
2. **Environment File**: `.env.dev` configured for development
3. **Backup File**: Timestamped backup of original database

To use the development database:

```bash
# Option 1: Set environment variable
export ENV_FILE=.env.dev
./bin/scheduler

# Option 2: Run with environment variable
ENV_FILE=.env.dev ./bin/scheduler

# Option 3: Modify the environment file path in your code
# The development server will run on port 8081 by default
```

## Security Notes

- **Sensitive Data**: Use the `-c` flag to clean sensitive data for development
- **Backups**: The scripts create backups - remember to clean them up when done
- **Passwords**: Development databases reset all passwords to "password"
- **Environment Files**: Don't commit `.env.dev` files with real credentials

## Troubleshooting

### Permission Errors
```bash
# Make sure scripts are executable
chmod +x scripts/*.sh
```

### Database Connection Issues
```bash
# Test your database connection first
mysql -u youruser -p -e "SELECT 1;"
```

### Missing .env File
```bash
# Create .env file with your database credentials
cp .env.example .env  # if you have an example file
# Edit .env with your actual credentials
```

### Port Conflicts
The development environment uses port 8081 by default. If this conflicts with other services, edit the `.env.dev` file and change the `SERVER_PORT` value.

## Clean Up

To remove development databases and files:

```bash
# Drop development database
mysql -u youruser -p -e "DROP DATABASE wmu_schedules_dev;"

# Remove environment file
rm .env.dev

# Remove backup files
rm /tmp/wmu_schedules_backup_*.sql
```
