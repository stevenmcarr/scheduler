#!/bin/bash
# Master deployment script for WMU Course Scheduler
# This script orchestrates the complete deployment process

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# Logging
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Show banner
show_banner() {
    echo -e "${PURPLE}"
    echo "============================================="
    echo "    WMU Course Scheduler Deployment"
    echo "============================================="
    echo -e "${NC}"
}

# Show help
show_help() {
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  deploy     - Full deployment (backup + compile + deploy)"
    echo "  quick      - Quick deployment (compile + deploy only)"
    echo "  backup     - Create backup only"
    echo "  restore    - Restore from backup"
    echo "  status     - Show deployment status"
    echo "  clean      - Remove deployment"
    echo "  logs       - Show service logs"
    echo "  help       - Show this help"
    echo ""
    echo "Options:"
    echo "  --no-backup     Skip backup creation"
    echo "  --no-restart    Skip service restart"
    echo "  --force         Force deployment even if service is running"
    echo ""
    echo "Examples:"
    echo "  $0 deploy               # Full deployment with backup"
    echo "  $0 quick                # Quick deployment for development"
    echo "  $0 deploy --no-backup   # Deploy without creating backup"
    echo "  $0 restore              # Show available backups"
    echo "  $0 logs                 # Show service logs"
}

# Parse arguments
parse_args() {
    COMMAND="${1:-deploy}"
    NO_BACKUP=false
    NO_RESTART=false
    FORCE=false
    
    shift || true
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-backup)
                NO_BACKUP=true
                shift
                ;;
            --no-restart)
                NO_RESTART=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            *)
                if [[ "$COMMAND" == "restore" ]]; then
                    RESTORE_FILE="$1"
                fi
                shift
                ;;
        esac
    done
}

# Execute deployment commands
execute_deploy() {
    log "Starting full deployment process..."
    
    # Create backup unless skipped
    if [[ "$NO_BACKUP" == "false" ]]; then
        log "Creating backup..."
        "$SCRIPT_DIR/backup-scheduler.sh" backup
    else
        warning "Skipping backup (--no-backup specified)"
    fi
    
    # Run deployment
    log "Running deployment..."
    "$SCRIPT_DIR/deploy-scheduler.sh" deploy
}

execute_quick() {
    log "Starting quick deployment..."
    "$SCRIPT_DIR/quick-deploy.sh"
}

execute_backup() {
    log "Creating backup..."
    "$SCRIPT_DIR/backup-scheduler.sh" backup
}

execute_restore() {
    if [[ -n "$RESTORE_FILE" ]]; then
        "$SCRIPT_DIR/backup-scheduler.sh" restore "$RESTORE_FILE"
    else
        log "Available backups:"
        "$SCRIPT_DIR/backup-scheduler.sh" list
        echo ""
        echo "Usage: $0 restore <backup_filename>"
    fi
}

execute_status() {
    "$SCRIPT_DIR/deploy-scheduler.sh" status
    echo ""
    echo "Service Status:"
    echo "==============="
    systemctl status scheduler.service --no-pager -l || true
}

execute_clean() {
    read -p "Are you sure you want to remove the deployment? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        "$SCRIPT_DIR/deploy-scheduler.sh" clean
    else
        log "Clean operation cancelled"
    fi
}

execute_logs() {
    echo "Recent logs (press Ctrl+C to exit):"
    echo "===================================="
    sudo journalctl -u scheduler.service -f --no-pager
}

# Main execution
main() {
    show_banner
    
    parse_args "$@"
    
    case "$COMMAND" in
        deploy)
            execute_deploy
            ;;
        quick)
            execute_quick
            ;;
        backup)
            execute_backup
            ;;
        restore)
            execute_restore
            ;;
        status)
            execute_status
            ;;
        clean)
            execute_clean
            ;;
        logs)
            execute_logs
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            error "Unknown command: $COMMAND"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
