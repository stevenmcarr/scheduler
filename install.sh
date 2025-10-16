#!/bin/bash
# WMU Course Scheduler Installation Script
# This script installs all software dependencies needed to run the WMU Course Scheduler
# Run this script after checking out the project from GitHub

set -e  # Exit on any error

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REQUIRED_GO_VERSION="1.23.0"
MYSQL_VERSION="8.0"

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
    echo "========================================="
    echo "  WMU Course Scheduler Installation"
    echo "========================================="
    echo -e "${NC}"
    echo "This script will install all required dependencies:"
    echo "• Go 1.23+ programming language"
    echo "• MySQL 8.0+ database server"
    echo "• Required Go packages"
    echo "• System log directories"
    echo "• TLS certificate directories"
    echo ""
}

# Detect Linux distribution
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        error "Cannot detect Linux distribution"
        exit 1
    fi
    
    log "Detected OS: $OS $VERSION"
}

# Check if running as root
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root or with sudo."
        error "Please run as your regular user account. The script will use sudo internally when needed."
        error "Usage: ./install.sh"
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
    
    success "Permission check passed"
}

# Update package manager
update_packages() {
    log "Updating package manager..."
    
    case $OS in
        ubuntu|debian)
            sudo apt update
            ;;
        centos|rhel|rocky|almalinux)
            sudo yum update -y || sudo dnf update -y
            ;;
        fedora)
            sudo dnf update -y
            ;;
        arch|manjaro)
            sudo pacman -Sy
            ;;
        *)
            warning "Unsupported distribution for automatic package updates: $OS"
            ;;
    esac
    
    success "Package manager updated"
}

# Install system dependencies
install_system_deps() {
    log "Installing system dependencies..."
    
    case $OS in
        ubuntu|debian)
            sudo apt install -y \
                curl \
                wget \
                git \
                build-essential \
                ca-certificates \
                gnupg \
                lsb-release \
                unzip \
                tar
            ;;
        centos|rhel|rocky|almalinux)
            sudo yum install -y \
                curl \
                wget \
                git \
                gcc \
                gcc-c++ \
                make \
                ca-certificates \
                gnupg \
                unzip \
                tar \
            || sudo dnf install -y \
                curl \
                wget \
                git \
                gcc \
                gcc-c++ \
                make \
                ca-certificates \
                gnupg \
                unzip \
                tar
            ;;
        fedora)
            sudo dnf install -y \
                curl \
                wget \
                git \
                gcc \
                gcc-c++ \
                make \
                ca-certificates \
                gnupg \
                unzip \
                tar
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm \
                curl \
                wget \
                git \
                base-devel \
                ca-certificates \
                gnupg \
                unzip \
                tar
            ;;
        *)
            warning "Unsupported distribution for automatic dependency installation: $OS"
            warning "Please manually install: curl wget git build-essential ca-certificates gnupg unzip tar"
            ;;
    esac
    
    success "System dependencies installed"
}

# Check Go version
check_go_version() {
    local version=$1
    local required_version=$2
    
    if [[ -z "$version" ]]; then
        return 1
    fi
    
    # Extract version numbers (e.g., "1.23.0" from "go version go1.23.0 linux/amd64")
    local current_version=$(echo "$version" | grep -oP 'go\K[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    
    if [[ -z "$current_version" ]]; then
        return 1
    fi
    
    # Simple version comparison
    local IFS='.'
    local current_parts=($current_version)
    local required_parts=($required_version)
    
    for i in {0..2}; do
        local current_part=${current_parts[$i]:-0}
        local required_part=${required_parts[$i]:-0}
        
        if [[ $current_part -gt $required_part ]]; then
            return 0
        elif [[ $current_part -lt $required_part ]]; then
            return 1
        fi
    done
    
    return 0
}

# Install Go
install_go() {
    log "Checking Go installation..."
    
    if command -v go &> /dev/null; then
        local go_version=$(go version 2>/dev/null || echo "")
        if check_go_version "$go_version" "$REQUIRED_GO_VERSION"; then
            success "Go is already installed: $go_version"
            return
        else
            warning "Go version is too old: $go_version"
            log "Required version: $REQUIRED_GO_VERSION or higher"
        fi
    fi
    
    log "Installing Go $REQUIRED_GO_VERSION..."
    
    # Determine architecture
    local arch=$(uname -m)
    case $arch in
        x86_64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        armv6l)
            arch="armv6l"
            ;;
        *)
            error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    # Download and install Go
    local go_version="1.23.11"  # Use latest stable from go.mod
    local go_tarball="go${go_version}.linux-${arch}.tar.gz"
    local go_url="https://go.dev/dl/${go_tarball}"
    
    log "Downloading Go from $go_url..."
    cd /tmp
    curl -LO "$go_url"
    
    log "Installing Go to /usr/local/go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$go_tarball"
    
    # Add Go to PATH
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    if ! grep -q "/usr/local/go/bin" ~/.profile 2>/dev/null; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    fi
    
    # Set Go path for current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Clean up
    rm -f "/tmp/$go_tarball"
    
    # Verify installation
    if command -v go &> /dev/null; then
        success "Go installed successfully: $(go version)"
    else
        error "Go installation failed"
        error "Please add /usr/local/go/bin to your PATH and restart your shell"
        exit 1
    fi
}

# Install MySQL
install_mysql() {
    log "Checking MySQL installation..."
    
    if command -v mysql &> /dev/null && systemctl is-active --quiet mysql 2>/dev/null; then
        success "MySQL is already installed and running"
        mysql --version
        return
    fi
    
    log "Installing MySQL Server..."
    
    case $OS in
        ubuntu|debian)
            # Install MySQL APT repository
            if ! dpkg -l | grep -q mysql-apt-config; then
                log "Adding MySQL APT repository..."
                cd /tmp
                wget https://dev.mysql.com/get/mysql-apt-config_0.8.29-1_all.deb
                sudo dpkg -i mysql-apt-config_0.8.29-1_all.deb || true
                sudo apt update
            fi
            
            # Install MySQL server
            sudo apt install -y mysql-server mysql-client
            ;;
        centos|rhel|rocky|almalinux)
            # Install MySQL repository
            if ! rpm -qa | grep -q mysql80-community-release; then
                log "Adding MySQL YUM repository..."
                sudo yum install -y https://dev.mysql.com/get/mysql80-community-release-el$(rpm -E %{rhel})-1.noarch.rpm || \
                sudo dnf install -y https://dev.mysql.com/get/mysql80-community-release-el$(rpm -E %{rhel})-1.noarch.rpm
            fi
            
            # Install MySQL server
            sudo yum install -y mysql-community-server mysql-community-client || \
            sudo dnf install -y mysql-community-server mysql-community-client
            ;;
        fedora)
            sudo dnf install -y mysql-server mysql
            ;;
        arch|manjaro)
            sudo pacman -S --noconfirm mysql
            ;;
        *)
            warning "Unsupported distribution for automatic MySQL installation: $OS"
            warning "Please manually install MySQL 8.0+ and ensure it's running"
            return
            ;;
    esac
    
    # Start and enable MySQL
    sudo systemctl start mysql || sudo systemctl start mysqld
    sudo systemctl enable mysql || sudo systemctl enable mysqld
    
    success "MySQL installed and started"
    
    # Secure MySQL installation prompt
    warning "IMPORTANT: Please run 'sudo mysql_secure_installation' to secure your MySQL installation"
    warning "Make sure to set a strong root password and remove test databases"
}

# Setup MySQL database and user
setup_database() {
    log "Setting up MySQL database and user..."
    
    # Check if MySQL is running
    if ! systemctl is-active --quiet mysql 2>/dev/null && ! systemctl is-active --quiet mysqld 2>/dev/null; then
        error "MySQL is not running. Please start MySQL and try again."
        exit 1
    fi
    
    echo ""
    info "Database setup requires MySQL root access."
    info "You will be prompted for the MySQL root password."
    echo ""
    
    # Create database and user
    mysql -u root -p << 'EOF'
-- Create databases
CREATE DATABASE IF NOT EXISTS wmu_schedules CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS wmu_schedules_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user
CREATE USER IF NOT EXISTS 'wmu_cs'@'localhost' IDENTIFIED BY '1h0ck3y$';

-- Grant privileges
GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'localhost';
GRANT ALL PRIVILEGES ON wmu_schedules_dev.* TO 'wmu_cs'@'localhost';

-- Flush privileges
FLUSH PRIVILEGES;

-- Show created databases
SHOW DATABASES LIKE 'wmu_schedules%';
EOF

    if [[ $? -eq 0 ]]; then
        success "Database and user created successfully"
    else
        error "Failed to create database and user"
        error "Please check your MySQL root password and try again"
        exit 1
    fi
}

# Install Go dependencies
install_go_deps() {
    log "Installing Go dependencies..."
    
    cd "$PROJECT_ROOT"
    
    # Initialize Go module if needed
    if [[ ! -f go.mod ]]; then
        warning "go.mod not found. This should not happen if you cloned the repository correctly."
        exit 1
    fi
    
    # Download dependencies
    log "Downloading Go modules..."
    go mod download
    
    # Verify dependencies
    log "Verifying Go modules..."
    go mod verify
    
    # Tidy up
    go mod tidy
    
    success "Go dependencies installed"
}

# Create required directories
create_directories() {
    log "Creating required directories..."
    
    # Create log directories
    sudo mkdir -p /var/log/scheduler
    sudo chown $USER:$USER /var/log/scheduler
    sudo chmod 755 /var/log/scheduler
    
    # Create uploads directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/src/uploads"
    mkdir -p "$PROJECT_ROOT/uploads"
    
    # Create certs directory for TLS certificates
    mkdir -p "$PROJECT_ROOT/certs"
    
    success "Required directories created"
}

# Setup environment file
setup_environment() {
    log "Setting up environment configuration..."
    
    cd "$PROJECT_ROOT"
    
    if [[ ! -f .env ]]; then
        if [[ -f .env.example ]]; then
            log "Copying .env.example to .env..."
            cp .env.example .env
            warning "Please edit .env file to configure your specific settings"
        else
            log "Creating default .env file..."
            cat > .env << 'EOF'
# Database Configuration
DB_USER=wmu_cs
DB_PASSWORD=1h0ck3y$
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=wmu_schedules

# Server Configuration
SERVER_PORT=4100

# TLS/HTTPS Configuration
TLS_ENABLED=false
# Optional: Specify custom certificate paths (leave empty for auto-detection)
TLS_CERT_FILE=
TLS_KEY_FILE=

# Logging Configuration
LOG_LEVEL=INFO
EOF
            warning "Created default .env file. Please review and modify as needed."
        fi
    else
        success ".env file already exists"
    fi
}

# Test installation
test_installation() {
    log "Testing installation..."
    
    cd "$PROJECT_ROOT"
    
    # Test Go build
    log "Testing Go build..."
    if go build -o scheduler-test ./src; then
        success "Go build successful"
        rm -f scheduler-test
    else
        error "Go build failed"
        exit 1
    fi
    
    # Test database connection (if configured)
    if [[ -f .env ]]; then
        log "Testing database connection..."
        # We'll create a simple test program
        cat > db_test.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }
    
    // Get database credentials
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")
    
    if dbUser == "" {
        dbUser = "wmu_cs"
    }
    if dbHost == "" {
        dbHost = "127.0.0.1"
    }
    if dbPort == "" {
        dbPort = "3306"
    }
    if dbName == "" {
        dbName = "wmu_schedules"
    }
    
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
        dbUser, dbPassword, dbHost, dbPort, dbName)
    
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }
    
    fmt.Println("Database connection successful!")
}
EOF
        
        if go run db_test.go 2>/dev/null; then
            success "Database connection test successful"
        else
            warning "Database connection test failed - please check your .env configuration"
        fi
        
        rm -f db_test.go
    fi
    
    success "Installation test completed"
}

# Print post-installation instructions
print_instructions() {
    echo ""
    echo -e "${CYAN}"
    echo "========================================="
    echo "     Installation Complete!"
    echo "========================================="
    echo -e "${NC}"
    echo ""
    echo "🎉 All dependencies have been installed successfully!"
    echo ""
    echo "📋 Next Steps:"
    echo ""
    echo "1. 📝 Configure your environment:"
    echo "   • Edit .env file with your specific settings"
    echo "   • Update database credentials if needed"
    echo ""
    echo "2. 🗄️  Set up database schema:"
    echo "   • Run database migration scripts in sql/ directory"
    echo "   • Import initial data if available"
    echo ""
    echo "3. 👤 Create admin user:"
    echo "   • Use ./add_user.sh or go run add_user.go"
    echo "   • Create at least one administrator account"
    echo ""
    echo "4. 🚀 Build and run the application:"
    echo "   • Development: go run ./src"
    echo "   • Production: use ./scripts/deploy-scheduler.sh"
    echo ""
    echo "📁 Important Files:"
    echo "   • .env - Environment configuration"
    echo "   • sql/ - Database schema and migrations"
    echo "   • src/ - Application source code"
    echo "   • scripts/ - Deployment and management scripts"
    echo ""
    echo "🔧 Useful Commands:"
    echo "   • Build: go build -o scheduler ./src"
    echo "   • Run: ./scheduler"
    echo "   • Test: go test ./..."
    echo "   • Deploy: ./scripts/deploy-scheduler.sh"
    echo ""
    echo "📖 Documentation:"
    echo "   • Check *.md files for specific setup guides"
    echo "   • ADD_USER_README.md - User management"
    echo "   • SECURITY.md - Security configuration"
    echo ""
    echo "🔍 Troubleshooting:"
    echo "   • Check logs: tail -f /var/log/scheduler/scheduler.log"
    echo "   • Verify services: systemctl status mysql"
    echo "   • Test database: mysql -u wmu_cs -p wmu_schedules"
    echo ""
    success "Happy coding! 🚀"
}

# Main installation function
main() {
    print_header
    
    detect_os
    check_permissions
    update_packages
    install_system_deps
    install_go
    install_mysql
    setup_database
    install_go_deps
    create_directories
    setup_environment
    test_installation
    print_instructions
}

# Handle command line arguments
case "${1:-install}" in
    install)
        main
        ;;
    deps-only)
        log "Installing dependencies only (no database setup)..."
        detect_os
        check_permissions
        update_packages
        install_system_deps
        install_go
        install_mysql
        install_go_deps
        create_directories
        setup_environment
        success "Dependencies installed (database setup skipped)"
        ;;
    test)
        log "Testing installation..."
        test_installation
        ;;
    *)
        echo "Usage: $0 {install|deps-only|test}"
        echo ""
        echo "Commands:"
        echo "  install    - Full installation including database setup (default)"
        echo "  deps-only  - Install dependencies only, skip database setup"
        echo "  test       - Test the current installation"
        echo ""
        echo "Run this script after cloning the repository from GitHub:"
        echo "  git clone https://github.com/stevenmcarr/scheduler.git"
        echo "  cd scheduler"
        echo "  ./install.sh"
        echo ""
        exit 1
        ;;
esac
