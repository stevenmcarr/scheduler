#!/bin/bash

# Simple wrapper script that reads database credentials from .env file
# and calls the main create_dev_database.sh script

set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Check if .env file exists
ENV_FILE="$PROJECT_DIR/.env"
if [[ ! -f "$ENV_FILE" ]]; then
    print_error ".env file not found at $ENV_FILE"
    print_info "Please create a .env file with your database credentials first."
    exit 1
fi

print_info "Reading database credentials from $ENV_FILE"

# Source the .env file to get variables
set -a  # Automatically export all variables
source "$ENV_FILE"
set +a  # Turn off automatic export

# Check if required variables are set
if [[ -z "$DB_USER" || -z "$DB_PASSWORD" ]]; then
    print_error "DB_USER and DB_PASSWORD must be set in .env file"
    exit 1
fi

# Set defaults for optional variables
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_NAME=${DB_NAME:-wmu_schedules}

print_info "Using database connection:"
print_info "  Host: $DB_HOST:$DB_PORT"
print_info "  User: $DB_USER"
print_info "  Source DB: $DB_NAME"

# Build arguments for the main script
ARGS="-u $DB_USER -p $DB_PASSWORD -H $DB_HOST -P $DB_PORT -s $DB_NAME"

# Check command line arguments to pass through
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            echo "Simple wrapper for create_dev_database.sh that reads credentials from .env"
            echo ""
            echo "This script will:"
            echo "  1. Read DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME from .env"
            echo "  2. Call create_dev_database.sh with these credentials"
            echo "  3. Pass through any additional arguments"
            echo ""
            echo "Additional options (passed to create_dev_database.sh):"
            echo "  -c, --clean-sensitive   Clean sensitive data from development database"
            echo "  -f, --force             Drop existing development database if it exists"
            echo "  -d, --dev-db DB         Development database name (default: wmu_schedules_dev)"
            echo ""
            echo "Examples:"
            echo "  $0                      # Basic copy"
            echo "  $0 -c -f               # Clean sensitive data and force overwrite"
            echo "  $0 -d my_dev_db        # Use custom dev database name"
            exit 0
            ;;
        *)
            ARGS="$ARGS $1"
            shift
            ;;
    esac
done

print_info "Calling create_dev_database.sh with arguments: $ARGS"
print_warning "This will create a development copy of your database."

# Ask for confirmation
echo -n "Continue? [y/N]: "
read -r confirm
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    print_info "Operation cancelled by user"
    exit 0
fi

# Call the main script
exec "$SCRIPT_DIR/create_dev_database.sh" $ARGS
