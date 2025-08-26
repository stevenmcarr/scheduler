#!/bin/bash

# WMU Scheduler - Add User Script Runner
# This script runs the Go add_user.go program

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to the scheduler directory
cd "$SCRIPT_DIR"

echo -e "${BLUE}WMU Scheduler - Add User Utility${NC}"
echo "================================="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    echo "Please install Go to use this script."
    exit 1
fi

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${RED}Error: .env file not found${NC}"
    echo "Please ensure .env file exists in the scheduler directory."
    exit 1
fi

# Check if add_user.go exists
if [ ! -f "add_user.go" ]; then
    echo -e "${RED}Error: add_user.go not found${NC}"
    echo "Please ensure add_user.go exists in the scheduler directory."
    exit 1
fi

echo -e "${YELLOW}Checking Go dependencies...${NC}"

# Check and install dependencies if needed
go mod tidy > /dev/null 2>&1

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to resolve Go dependencies${NC}"
    echo "Please check your Go module setup."
    exit 1
fi

echo -e "${GREEN}Dependencies OK${NC}"
echo

# Run the add_user.go script
echo -e "${YELLOW}Starting add user process...${NC}"
echo

go run add_user.go

# Check if the script ran successfully
if [ $? -eq 0 ]; then
    echo
    echo -e "${GREEN}Add user script completed successfully!${NC}"
else
    echo
    echo -e "${RED}Add user script failed with an error.${NC}"
    exit 1
fi
