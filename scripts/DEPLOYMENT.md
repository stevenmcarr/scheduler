# WMU Course Scheduler Deployment Guide

This directory contains scripts for compiling and deploying the WMU Course Scheduler application to `/var/www/html/scheduler`.

## Quick Start

For most deployments, simply run:
```bash
./scripts/manage-deployment.sh deploy
```

For development iterations:
```bash
./scripts/manage-deployment.sh quick
```

## Scripts Overview

### 1. manage-deployment.sh (Master Script)
The main deployment orchestrator that provides a unified interface for all deployment operations.

**Usage:**
```bash
./scripts/manage-deployment.sh [COMMAND] [OPTIONS]
```

**Commands:**
- `deploy` - Full deployment with backup and compilation
- `quick` - Quick deployment for development (no backup)
- `backup` - Create backup only
- `restore` - Restore from backup
- `status` - Show deployment status
- `clean` - Remove deployment completely
- `logs` - Show real-time service logs

**Options:**
- `--no-backup` - Skip backup creation
- `--no-restart` - Skip service restart
- `--force` - Force deployment even if service is running

### 2. deploy-scheduler.sh (Full Deployment)
Comprehensive deployment script that handles:
- Prerequisites checking
- Application compilation
- File deployment
- Service configuration
- Security hardening

### 3. quick-deploy.sh (Development)
Lightweight script for rapid deployment during development:
- Quick compilation
- Template updates
- Service restart
- Basic status check

### 4. backup-scheduler.sh (Backup Management)
Handles backup and restore operations:
- Creates compressed backups
- Manages backup retention
- Provides restore functionality
- Lists available backups

## Deployment Process

### Full Deployment Flow
1. **Prerequisites Check** - Verifies Go installation, project structure
2. **Service Stop** - Gracefully stops the running service
3. **Backup Creation** - Creates timestamped backup of current deployment
4. **Compilation** - Builds optimized binary with version info
5. **File Deployment** - Copies binary, templates, assets, certificates
6. **Permission Setup** - Sets proper ownership and security permissions
7. **Service Configuration** - Updates systemd service file
8. **Service Start** - Starts and enables the service
9. **Health Check** - Verifies service is responding

### Directory Structure
```
/var/www/html/scheduler/
├── scheduler              # Main binary
├── templates/            # Go HTML templates
├── images/              # Static images
├── uploads/             # User uploaded files
├── certs/               # TLS certificates
├── logs/                # Application logs
└── .env                 # Environment configuration
```

## Usage Examples

### Development Workflow
```bash
# Make changes to code
vim src/main.go

# Quick deploy for testing
./scripts/manage-deployment.sh quick

# Check if working
curl http://localhost:8080/scheduler/
```

### Production Deployment
```bash
# Full deployment with backup
./scripts/manage-deployment.sh deploy

# Check status
./scripts/manage-deployment.sh status

# Monitor logs
./scripts/manage-deployment.sh logs
```

### Backup and Restore
```bash
# Create backup
./scripts/manage-deployment.sh backup

# List backups
./scripts/backup-scheduler.sh list

# Restore specific backup
./scripts/manage-deployment.sh restore scheduler_backup_20240728_143022.tar.gz
```

### Troubleshooting
```bash
# Check service status
systemctl status scheduler.service

# View detailed logs
sudo journalctl -u scheduler.service -f

# Check deployment status
./scripts/manage-deployment.sh status

# Manual service control
sudo systemctl restart scheduler.service
```

## Configuration

### Environment Variables
The application looks for configuration in this order:
1. `.env` file in deployment directory
2. `.env.debug` for development
3. System environment variables

### Service Configuration
The systemd service is configured with:
- **User/Group**: www-data
- **Security**: NoNewPrivileges, ProtectSystem, PrivateTmp
- **Restart**: Always with 10-second delay
- **Logging**: systemd journal integration

---

For additional help or issues, check the service logs or run the status command to diagnose deployment problems.
