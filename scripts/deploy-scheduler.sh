#!/bin/bash
# WMU Course Scheduler Deployment Script
# This script compiles and deploys the scheduler application to /var/www/html/scheduler

set -e  # Exit on any error

# Configuration
PROJECT_ROOT="/home/stevecarr/scheduler"
DEPLOY_DIR="/var/www/html/scheduler"
SERVICE_NAME="scheduler.service"
BINARY_NAME="scheduler"
BUILD_DIR="$PROJECT_ROOT/build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
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

# Helper function to run sudo commands while preserving user environment
run_sudo() {
    if [[ $EUID -eq 0 ]]; then
        # Already root, just run the command
        "$@"
    else
        # Use sudo with preserved environment where needed
        sudo "$@"
    fi
}

# Check if running as root or with sudo
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root or with sudo."
        error "Please run as your regular user account. The script will use sudo internally when needed."
        error "Usage: ./deploy-scheduler.sh"
        exit 1
    fi
    
    # Check if user can sudo
    if ! sudo -n true 2>/dev/null; then
        log "Testing sudo access..."
        if ! sudo true; then
            error "This script requires sudo access. Please ensure you can run sudo commands."
            exit 1
        fi
    fi
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        error "Go is not installed or not in PATH"
        exit 1
    fi
    
    log "Go version: $(go version)"
    
    # Check if project directory exists
    if [[ ! -d "$PROJECT_ROOT" ]]; then
        error "Project directory not found: $PROJECT_ROOT"
        exit 1
    fi
    
    # Check if source files exist
    if [[ ! -f "$PROJECT_ROOT/src/main.go" ]]; then
        error "main.go not found in $PROJECT_ROOT/src/"
        exit 1
    fi
    
    success "Prerequisites check passed"
}

# Stop the service if it's running
stop_service() {
    log "Checking if service is running..."
    if systemctl is-active --quiet $SERVICE_NAME 2>/dev/null; then
        log "Stopping $SERVICE_NAME..."
        run_sudo systemctl stop $SERVICE_NAME
        success "Service stopped"
    else
        log "Service is not running"
    fi
}

# Create build directory
prepare_build() {
    log "Preparing build environment..."
    
    # Remove old build directory
    if [[ -d "$BUILD_DIR" ]]; then
        rm -rf "$BUILD_DIR"
    fi
    
    # Create fresh build directory
    mkdir -p "$BUILD_DIR"
    success "Build directory prepared"
}

# Compile the application
compile_application() {
    log "Compiling application..."
    
    cd "$PROJECT_ROOT"
    
    # Set build environment
    export CGO_ENABLED=1
    export GOOS=linux
    export GOARCH=amd64
    
    # Build the application
    log "Building binary..."
    go build -o "$BUILD_DIR/$BINARY_NAME" \
        -ldflags "-X main.version=$(date +%Y%m%d-%H%M%S) -w -s" \
        ./src
    
    if [[ ! -f "$BUILD_DIR/$BINARY_NAME" ]]; then
        error "Failed to build binary"
        exit 1
    fi
    
    # Make binary executable
    chmod +x "$BUILD_DIR/$BINARY_NAME"
    
    success "Application compiled successfully"
    log "Binary size: $(du -h "$BUILD_DIR/$BINARY_NAME" | cut -f1)"
}

# Prepare deployment directory
prepare_deployment() {
    log "Preparing deployment directory..."
    
    # Create deployment directory if it doesn't exist
    run_sudo mkdir -p "$DEPLOY_DIR"
    
    # Create subdirectories
    run_sudo mkdir -p "$DEPLOY_DIR/templates"
    run_sudo mkdir -p "$DEPLOY_DIR/images"
    run_sudo mkdir -p "$DEPLOY_DIR/uploads"
    run_sudo mkdir -p "$DEPLOY_DIR/certs"
    run_sudo mkdir -p "$DEPLOY_DIR/logs"
    
    success "Deployment directory prepared"
}

# Deploy files
deploy_files() {
    log "Deploying files..."
    
    # Deploy binary
    log "Deploying binary..."
    run_sudo cp "$BUILD_DIR/$BINARY_NAME" "$DEPLOY_DIR/"
    run_sudo chmod +x "$DEPLOY_DIR/$BINARY_NAME"
    
    # Deploy templates
    log "Deploying templates..."
    run_sudo cp -r "$PROJECT_ROOT/src/templates/"* "$DEPLOY_DIR/templates/"
    
    # Deploy static assets
    log "Deploying static assets..."
    if [[ -d "$PROJECT_ROOT/images" ]]; then
        run_sudo cp -r "$PROJECT_ROOT/images/"* "$DEPLOY_DIR/images/" 2>/dev/null || true
    fi
    
    # Deploy configuration files
    log "Deploying configuration..."
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        run_sudo cp "$PROJECT_ROOT/.env" "$DEPLOY_DIR/"
    elif [[ -f "$PROJECT_ROOT/.env.example" ]]; then
        warning ".env not found, copying .env.example"
        run_sudo cp "$PROJECT_ROOT/.env.example" "$DEPLOY_DIR/.env"
    else
        warning "No environment file found"
    fi
    
    # Deploy certificates if they exist
    if [[ -d "$PROJECT_ROOT/certs" ]]; then
        log "Deploying certificates..."
        run_sudo cp -r "$PROJECT_ROOT/certs/"* "$DEPLOY_DIR/certs/" 2>/dev/null || true
    fi
    
    # Set proper ownership and permissions
    log "Setting permissions..."
    run_sudo chown -R www-data:www-data "$DEPLOY_DIR"
    run_sudo chmod -R 755 "$DEPLOY_DIR"
    run_sudo chmod +x "$DEPLOY_DIR/$BINARY_NAME"
    
    # Ensure log directory is writable
    run_sudo chmod 755 "$DEPLOY_DIR/logs"
    
    success "Files deployed successfully"
}

# Update systemd service file
update_service() {
    log "Updating systemd service..."
    
    # Create system log directory for scheduler
    log "Creating system log directory..."
    sudo mkdir -p /var/log/scheduler
    sudo chown www-data:www-data /var/log/scheduler
    sudo chmod 755 /var/log/scheduler
    
    SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME"
    
    # Create service file
    sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=WMU Course Scheduler
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/$BINARY_NAME
Restart=always
RestartSec=10
Environment=GIN_MODE=release
StandardOutput=journal
StandardError=journal
SyslogIdentifier=scheduler

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$DEPLOY_DIR/logs $DEPLOY_DIR/uploads /var/log/scheduler
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    sudo systemctl daemon-reload
    success "Service file updated"
}

# Start the service
start_service() {
    log "Starting service..."
    
    # Enable service to start on boot
    sudo systemctl enable $SERVICE_NAME
    
    # Start the service
    sudo systemctl start $SERVICE_NAME
    
    # Wait a moment for service to start
    sleep 3
    
    # Check service status
    if systemctl is-active --quiet $SERVICE_NAME; then
        success "Service started successfully"
        
        # Show service status
        sudo systemctl status $SERVICE_NAME --no-pager -l
        
        # Test connectivity
        log "Testing connectivity..."
        sleep 2
        
        # Try both HTTP and HTTPS
        for protocol in http https; do
            for port in 8080 8081 4100; do
                log "Testing $protocol://localhost:$port/scheduler/"
                if curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 "$protocol://localhost:$port/scheduler/" 2>/dev/null | grep -q "200\|302\|404"; then
                    success "Service responding on $protocol://localhost:$port"
                    break 2
                fi
            done
        done
    else
        error "Failed to start service"
        log "Service logs:"
        sudo journalctl -u $SERVICE_NAME --no-pager -n 20
        exit 1
    fi
}

# Cleanup build directory
cleanup() {
    log "Cleaning up..."
    if [[ -d "$BUILD_DIR" ]]; then
        rm -rf "$BUILD_DIR"
    fi
    success "Cleanup completed"
}

# Show deployment info
show_deployment_info() {
    echo ""
    echo "================================================="
    echo "         DEPLOYMENT COMPLETED SUCCESSFULLY"
    echo "================================================="
    echo ""
    echo "🏠 Deployment Directory: $DEPLOY_DIR"
    echo "📁 Binary Location: $DEPLOY_DIR/$BINARY_NAME"
    echo "🔧 Service Name: $SERVICE_NAME"
    echo "📊 Service Status: $(systemctl is-active $SERVICE_NAME)"
    echo "🚀 Auto-start: $(systemctl is-enabled $SERVICE_NAME)"
    echo ""
    echo "Management Commands:"
    echo "  Status:  sudo systemctl status $SERVICE_NAME"
    echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
    echo "  Start:   sudo systemctl start $SERVICE_NAME"
    echo "  Restart: sudo systemctl restart $SERVICE_NAME"
    echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
    echo ""
    echo "🌐 Test URLs:"
    echo "  HTTP:  http://localhost:8080/scheduler/"
    echo "  HTTPS: https://localhost:8081/scheduler/"
    echo ""
}

# Main execution
main() {
    log "Starting WMU Course Scheduler deployment..."
    echo ""
    
    check_permissions
    check_prerequisites
    stop_service
    prepare_build
    compile_application
    prepare_deployment
    deploy_files
    update_service
    start_service
    cleanup
    show_deployment_info
    
    success "Deployment completed successfully! 🎉"
}

# Handle script arguments
case "${1:-deploy}" in
    deploy)
        main
        ;;
    clean)
        log "Cleaning deployment..."
        stop_service
        sudo rm -rf "$DEPLOY_DIR"
        sudo systemctl disable $SERVICE_NAME 2>/dev/null || true
        sudo rm -f "/etc/systemd/system/$SERVICE_NAME"
        sudo systemctl daemon-reload
        success "Deployment cleaned"
        ;;
    status)
        echo "Deployment Status:"
        echo "=================="
        echo "Deploy Dir: $DEPLOY_DIR"
        echo "Binary: $(ls -la "$DEPLOY_DIR/$BINARY_NAME" 2>/dev/null || echo "Not found")"
        echo "Service: $(systemctl is-active $SERVICE_NAME 2>/dev/null || echo "Not active")"
        if [[ -f "$DEPLOY_DIR/$BINARY_NAME" ]]; then
            echo "Binary Info: $("$DEPLOY_DIR/$BINARY_NAME" --version 2>/dev/null || echo "No version info")"
        fi
        ;;
    *)
        echo "Usage: $0 {deploy|clean|status}"
        echo ""
        echo "Commands:"
        echo "  deploy  - Compile and deploy the application (default)"
        echo "  clean   - Remove deployment and stop service"
        echo "  status  - Show deployment status"
        echo ""
        echo "IMPORTANT: Run this script as your regular user, NOT with sudo:"
        echo "  Correct:   ./deploy-scheduler.sh"
        echo "  Incorrect: sudo ./deploy-scheduler.sh"
        echo ""
        echo "The script will use sudo internally when needed."
        exit 1
        ;;
esac
