#!/bin/bash
# WMU Course Scheduler Database Copy Script
# This script creates a backup of the local wmu_schedules database and copies it to vaccs.cs.wmich.edu

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="$PROJECT_ROOT/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="wmu_schedules_backup_${TIMESTAMP}.sql"
BACKUP_PATH="$BACKUP_DIR/$BACKUP_FILE"

# Remote server configuration
REMOTE_HOST="vaccs.cs.wmich.edu"
REMOTE_USER="${REMOTE_USER:-$USER}"  # Use current user if not specified
REMOTE_PATH="/tmp"
REMOTE_BACKUP_FILE="$REMOTE_PATH/$BACKUP_FILE"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

# Print header
print_header() {
    echo -e "${CYAN}"
    echo "=================================================="
    echo "  WMU Scheduler Database Copy to VACCS"
    echo "=================================================="
    echo -e "${NC}"
    echo "This script will:"
    echo "• Create a backup of the local wmu_schedules database"
    echo "• Copy the backup to vaccs.cs.wmich.edu"
    echo "• Optionally restore the database on the remote server"
    echo ""
}

# Load environment variables
load_env() {
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        log "Loading environment variables from .env..."
        source "$PROJECT_ROOT/.env"
    else
        warning ".env file not found. Using default values."
        DB_USER="wmu_cs"
        DB_PASSWORD="1h0ck3y$"
        DB_HOST="127.0.0.1"
        DB_PORT="3306"
        DB_NAME="wmu_schedules"
    fi
    
    info "Database: $DB_NAME on $DB_HOST:$DB_PORT"
    info "Remote host: $REMOTE_HOST"
    info "Remote user: $REMOTE_USER"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if MySQL client is installed
    if ! command -v mysql &> /dev/null; then
        error "MySQL client is not installed. Please install mysql-client package."
        exit 1
    fi
    
    if ! command -v mysqldump &> /dev/null; then
        error "mysqldump is not installed. Please install mysql-client package."
        exit 1
    fi
    
    # Check if ssh is available
    if ! command -v ssh &> /dev/null; then
        error "SSH client is not installed. Please install openssh-client package."
        exit 1
    fi
    
    if ! command -v scp &> /dev/null; then
        error "SCP is not installed. Please install openssh-client package."
        exit 1
    fi
    
    # Test local database connection
    log "Testing local database connection..."
    if ! mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME; SELECT 1;" &>/dev/null; then
        error "Cannot connect to local database. Please check your credentials."
        exit 1
    fi
    
    # Test SSH connection to remote host
    log "Testing SSH connection to $REMOTE_HOST..."
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes "$REMOTE_USER@$REMOTE_HOST" exit 2>/dev/null; then
        warning "Cannot connect to $REMOTE_HOST using SSH keys."
        warning "You will be prompted for password during file transfer."
        echo ""
        info "To set up SSH key authentication:"
        info "1. Generate SSH key: ssh-keygen -t rsa -b 4096"
        info "2. Copy to remote: ssh-copy-id $REMOTE_USER@$REMOTE_HOST"
        echo ""
        read -p "Do you want to continue with password authentication? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Operation cancelled."
            exit 0
        fi
    else
        success "SSH connection to $REMOTE_HOST successful"
    fi
    
    success "Prerequisites check passed"
}

# Create backup directory
create_backup_dir() {
    log "Creating backup directory..."
    mkdir -p "$BACKUP_DIR"
    success "Backup directory: $BACKUP_DIR"
}

# Create database backup
create_backup() {
    log "Creating database backup..."
    
    # Get database size for progress indication
    local db_size=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" \
        -e "SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 1) AS 'DB Size in MB' 
            FROM information_schema.tables WHERE table_schema='$DB_NAME';" -s -N 2>/dev/null || echo "Unknown")
    
    info "Database size: ${db_size} MB"
    info "Creating backup: $BACKUP_PATH"
    
    # Create mysqldump with comprehensive options
    mysqldump \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --user="$DB_USER" \
        --password="$DB_PASSWORD" \
        --single-transaction \
        --routines \
        --triggers \
        --events \
        --add-drop-database \
        --create-options \
        --disable-keys \
        --extended-insert \
        --quick \
        --lock-tables=false \
        --set-gtid-purged=OFF \
        --default-character-set=utf8mb4 \
        --databases "$DB_NAME" \
        > "$BACKUP_PATH"
    
    if [[ $? -eq 0 ]]; then
        local backup_size=$(du -h "$BACKUP_PATH" | cut -f1)
        success "Database backup created successfully: $backup_size"
    else
        error "Failed to create database backup"
        exit 1
    fi
}

# Copy backup to remote server
copy_to_remote() {
    log "Copying backup to $REMOTE_HOST..."
    
    local backup_size=$(du -h "$BACKUP_PATH" | cut -f1)
    info "Transferring $backup_size to $REMOTE_HOST:$REMOTE_BACKUP_FILE"
    
    # Use scp to copy the file
    if scp "$BACKUP_PATH" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_BACKUP_FILE"; then
        success "Backup copied to remote server successfully"
    else
        error "Failed to copy backup to remote server"
        exit 1
    fi
}

# Verify remote backup
verify_remote_backup() {
    log "Verifying remote backup..."
    
    # Check if file exists and get its size
    local remote_size=$(ssh "$REMOTE_USER@$REMOTE_HOST" "ls -lh '$REMOTE_BACKUP_FILE' 2>/dev/null | awk '{print \$5}'" 2>/dev/null || echo "Error")
    
    if [[ "$remote_size" != "Error" ]]; then
        success "Remote backup verified: $remote_size"
        
        # Get checksums for verification
        local local_checksum=$(md5sum "$BACKUP_PATH" | awk '{print $1}')
        local remote_checksum=$(ssh "$REMOTE_USER@$REMOTE_HOST" "md5sum '$REMOTE_BACKUP_FILE' 2>/dev/null | awk '{print \$1}'" 2>/dev/null || echo "Error")
        
        if [[ "$local_checksum" == "$remote_checksum" && "$remote_checksum" != "Error" ]]; then
            success "Checksum verification passed: $local_checksum"
        else
            warning "Checksum verification failed or unavailable"
            warning "Local: $local_checksum"
            warning "Remote: $remote_checksum"
        fi
    else
        error "Remote backup verification failed"
        exit 1
    fi
}

# Restore database on remote server (optional)
restore_on_remote() {
    echo ""
    info "The backup has been successfully copied to $REMOTE_HOST"
    info "Remote backup location: $REMOTE_BACKUP_FILE"
    echo ""
    
    read -p "Do you want to restore the database on the remote server now? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "Restoring database on remote server..."
        
        echo ""
        info "You will need to provide MySQL credentials for the remote server."
        read -p "Remote MySQL username [wmu_cs]: " remote_db_user
        remote_db_user=${remote_db_user:-wmu_cs}
        
        read -p "Remote MySQL host [127.0.0.1]: " remote_db_host
        remote_db_host=${remote_db_host:-127.0.0.1}
        
        read -p "Remote MySQL port [3306]: " remote_db_port
        remote_db_port=${remote_db_port:-3306}
        
        # Execute restore on remote server
        ssh -t "$REMOTE_USER@$REMOTE_HOST" << EOF
echo "Restoring database on remote server..."
echo "You will be prompted for the MySQL password."
echo ""

# Check if MySQL client is available
if ! command -v mysql &> /dev/null; then
    echo "Error: MySQL client is not installed on the remote server."
    exit 1
fi

# Restore the database
mysql -h "$remote_db_host" -P "$remote_db_port" -u "$remote_db_user" -p < "$REMOTE_BACKUP_FILE"

if [ \$? -eq 0 ]; then
    echo "Database restored successfully on remote server!"
    
    # Clean up the backup file
    read -p "Remove the backup file from remote server? (Y/n): " -n 1 -r
    echo
    if [[ ! \$REPLY =~ ^[Nn]$ ]]; then
        rm -f "$REMOTE_BACKUP_FILE"
        echo "Backup file removed from remote server."
    fi
else
    echo "Error: Database restore failed."
    exit 1
fi
EOF
        
        if [[ $? -eq 0 ]]; then
            success "Database restored successfully on remote server"
        else
            error "Database restore failed on remote server"
            warning "Backup file remains at: $REMOTE_BACKUP_FILE"
        fi
    else
        info "Database restore skipped."
        info "To restore manually on the remote server, run:"
        info "  ssh $REMOTE_USER@$REMOTE_HOST"
        info "  mysql -u username -p < $REMOTE_BACKUP_FILE"
    fi
}

# Clean up local backup (optional)
cleanup_local() {
    echo ""
    read -p "Remove the local backup file? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f "$BACKUP_PATH"
        success "Local backup file removed"
    else
        info "Local backup preserved at: $BACKUP_PATH"
    fi
}

# Print completion summary
print_summary() {
    echo ""
    echo -e "${CYAN}"
    echo "=================================================="
    echo "           Transfer Complete!"
    echo "=================================================="
    echo -e "${NC}"
    echo ""
    success "Database backup and transfer completed successfully!"
    echo ""
    echo "📋 Summary:"
    echo "  • Source Database: $DB_NAME on $DB_HOST"
    echo "  • Backup Created: $TIMESTAMP"
    echo "  • Remote Server: $REMOTE_HOST"
    echo "  • Remote User: $REMOTE_USER"
    echo ""
    echo "🔧 Remote Commands:"
    echo "  • Connect: ssh $REMOTE_USER@$REMOTE_HOST"
    echo "  • Restore: mysql -u username -p < $REMOTE_BACKUP_FILE"
    echo "  • Verify: mysql -u username -p -e 'SHOW DATABASES;'"
    echo ""
    echo "📁 Files:"
    if [[ -f "$BACKUP_PATH" ]]; then
        echo "  • Local backup: $BACKUP_PATH"
    fi
    echo "  • Remote backup: $REMOTE_BACKUP_FILE"
    echo ""
}

# Main function
main() {
    print_header
    load_env
    check_prerequisites
    create_backup_dir
    create_backup
    copy_to_remote
    verify_remote_backup
    restore_on_remote
    cleanup_local
    print_summary
}

# Handle command line arguments
case "${1:-copy}" in
    copy)
        main
        ;;
    backup-only)
        log "Creating backup only (no remote transfer)..."
        load_env
        check_prerequisites
        create_backup_dir
        create_backup
        success "Backup created at: $BACKUP_PATH"
        ;;
    transfer-only)
        if [[ -z "$2" ]]; then
            error "Usage: $0 transfer-only <backup-file>"
            error "Example: $0 transfer-only /path/to/backup.sql"
            exit 1
        fi
        BACKUP_PATH="$2"
        BACKUP_FILE=$(basename "$BACKUP_PATH")
        REMOTE_BACKUP_FILE="$REMOTE_PATH/$BACKUP_FILE"
        
        if [[ ! -f "$BACKUP_PATH" ]]; then
            error "Backup file not found: $BACKUP_PATH"
            exit 1
        fi
        
        load_env
        copy_to_remote
        verify_remote_backup
        restore_on_remote
        success "Transfer completed"
        ;;
    *)
        echo "WMU Scheduler Database Copy Script"
        echo ""
        echo "Usage: $0 {copy|backup-only|transfer-only <file>}"
        echo ""
        echo "Commands:"
        echo "  copy           - Create backup and copy to remote server (default)"
        echo "  backup-only    - Create local backup only"
        echo "  transfer-only  - Transfer existing backup file to remote server"
        echo ""
        echo "Environment Variables:"
        echo "  REMOTE_USER    - Remote username (default: current user)"
        echo ""
        echo "Examples:"
        echo "  $0                                    # Full copy process"
        echo "  $0 backup-only                       # Backup only"
        echo "  $0 transfer-only backup.sql          # Transfer existing backup"
        echo "  REMOTE_USER=username $0              # Use specific remote user"
        echo ""
        exit 1
        ;;
esac