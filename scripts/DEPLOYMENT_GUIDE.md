# WMU Scheduler Deployment Guide

## How to Deploy

### Correct Usage
```bash
# Navigate to the scheduler directory
cd /home/stevecarr/scheduler

# Run the deployment script as your regular user (NOT with sudo)
./scripts/deploy-scheduler.sh
```

### What NOT to do
```bash
# ❌ DO NOT run with sudo
sudo ./scripts/deploy-scheduler.sh

# ❌ DO NOT run as root
su - root
./scripts/deploy-scheduler.sh
```

## How it Works

The deployment script:
1. **Runs as your regular user** - This ensures proper access to your project files
2. **Uses sudo internally** - When it needs elevated privileges for system operations
3. **Handles permissions correctly** - Maintains proper file ownership and access

## What the Script Does

1. **Checks prerequisites** - Verifies Go installation and project structure
2. **Stops existing service** - Safely stops the running scheduler service
3. **Builds the application** - Compiles the Go binary with optimizations
4. **Deploys files** - Copies files to `/var/www/html/scheduler/`
5. **Updates systemd service** - Creates/updates the service configuration
6. **Starts the service** - Launches the new version

## Commands Available

```bash
# Deploy the application (default)
./scripts/deploy-scheduler.sh
./scripts/deploy-scheduler.sh deploy

# Clean up deployment (removes files and stops service)
./scripts/deploy-scheduler.sh clean

# Check deployment status
./scripts/deploy-scheduler.sh status
```

## After Deployment

The service will be available at:
- HTTP: `http://localhost:8080/scheduler/`
- HTTPS: `https://localhost:8081/scheduler/`

## Managing the Service

```bash
# Check status
sudo systemctl status scheduler.service

# View logs
sudo journalctl -u scheduler.service -f

# Restart service
sudo systemctl restart scheduler.service

# Stop service
sudo systemctl stop scheduler.service

# Start service
sudo systemctl start scheduler.service
```

## Troubleshooting

### Permission Denied Error
If you see "This script should not be run as root", make sure you're running as your regular user:
```bash
# Check current user
whoami

# Should show your username (e.g., stevecarr), not root
```

### Sudo Password Prompt
The script will prompt for your sudo password when needed for system operations. This is normal and expected.

### Service Won't Start
Check the logs for details:
```bash
sudo journalctl -u scheduler.service --no-pager -n 50
```
