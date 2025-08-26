#!/bin/bash
# Backup script for WMU Course Scheduler
# Creates backups before deployment

set -e

DEPLOY_DIR="/var/www/html/scheduler"
BACKUP_DIR="/var/backups/scheduler"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="scheduler_backup_$TIMESTAMP"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Create backup directory
create_backup_dir() {
    sudo mkdir -p "$BACKUP_DIR"
    log "Backup directory: $BACKUP_DIR"
}

# Create backup
create_backup() {
    if [[ ! -d "$DEPLOY_DIR" ]]; then
        log "No deployment found at $DEPLOY_DIR - nothing to backup"
        return 0
    fi
    
    log "Creating backup: $BACKUP_NAME"
    
    # Create backup archive
    sudo tar -czf "$BACKUP_DIR/$BACKUP_NAME.tar.gz" \
        -C "$(dirname "$DEPLOY_DIR")" \
        "$(basename "$DEPLOY_DIR")" \
        --exclude="logs/*" \
        --exclude="uploads/temp/*" 2>/dev/null || true
    
    if [[ -f "$BACKUP_DIR/$BACKUP_NAME.tar.gz" ]]; then
        local size=$(du -h "$BACKUP_DIR/$BACKUP_NAME.tar.gz" | cut -f1)
        success "Backup created: $BACKUP_NAME.tar.gz ($size)"
    else
        log "Backup creation failed or deployment not found"
        return 1
    fi
}

# Clean old backups
clean_old_backups() {
    local retention_days=${1:-7}
    log "Cleaning backups older than $retention_days days..."
    
    sudo find "$BACKUP_DIR" -name "scheduler_backup_*.tar.gz" \
        -type f -mtime +$retention_days -delete 2>/dev/null || true
    
    local count=$(ls -1 "$BACKUP_DIR"/scheduler_backup_*.tar.gz 2>/dev/null | wc -l)
    log "Remaining backups: $count"
}

# List backups
list_backups() {
    echo "Available backups:"
    echo "=================="
    
    if [[ -d "$BACKUP_DIR" ]]; then
        sudo ls -lh "$BACKUP_DIR"/scheduler_backup_*.tar.gz 2>/dev/null | \
            awk '{print $9, $5, $6, $7, $8}' | \
            sed 's|.*/||' || echo "No backups found"
    else
        echo "Backup directory does not exist"
    fi
}

# Restore backup
restore_backup() {
    local backup_file="$1"
    
    if [[ -z "$backup_file" ]]; then
        echo "Usage: $0 restore <backup_filename>"
        echo ""
        list_backups
        return 1
    fi
    
    local backup_path="$BACKUP_DIR/$backup_file"
    if [[ ! -f "$backup_path" ]]; then
        backup_path="$BACKUP_DIR/${backup_file}.tar.gz"
    fi
    
    if [[ ! -f "$backup_path" ]]; then
        echo "Backup file not found: $backup_file"
        list_backups
        return 1
    fi
    
    log "Stopping service..."
    sudo systemctl stop scheduler.service 2>/dev/null || true
    
    log "Restoring from: $(basename "$backup_path")"
    
    # Backup current deployment
    if [[ -d "$DEPLOY_DIR" ]]; then
        log "Backing up current deployment..."
        sudo mv "$DEPLOY_DIR" "${DEPLOY_DIR}.pre-restore.$(date +%s)"
    fi
    
    # Restore from backup
    sudo mkdir -p "$(dirname "$DEPLOY_DIR")"
    sudo tar -xzf "$backup_path" -C "$(dirname "$DEPLOY_DIR")"
    
    # Set permissions
    sudo chown -R www-data:www-data "$DEPLOY_DIR"
    sudo chmod +x "$DEPLOY_DIR/scheduler"
    
    log "Starting service..."
    sudo systemctl start scheduler.service
    
    success "Restore completed"
}

# Main execution
case "${1:-backup}" in
    backup)
        create_backup_dir
        create_backup
        clean_old_backups 7
        ;;
    list)
        list_backups
        ;;
    restore)
        restore_backup "$2"
        ;;
    clean)
        retention_days=${2:-7}
        clean_old_backups "$retention_days"
        ;;
    *)
        echo "Usage: $0 {backup|list|restore|clean}"
        echo ""
        echo "Commands:"
        echo "  backup           - Create a new backup (default)"
        echo "  list             - List available backups"
        echo "  restore <file>   - Restore from backup file"
        echo "  clean [days]     - Clean backups older than N days (default: 7)"
        exit 1
        ;;
esac
