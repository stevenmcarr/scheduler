#!/bin/bash

# WMU Course Scheduler Unit Test Runner
# This script runs the complete unit test suite for the controller functions

echo "========================================"
echo "WMU Course Scheduler Unit Test Suite"
echo "========================================"
echo ""

# Change to the project root directory
cd "$(dirname "$0")/.."

# Set Go environment
export CGO_ENABLED=1
export GOOS=darwin
export GOARCH=amd64

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running unit tests for controller functions...${NC}"
echo ""

# Test 1: Authentication Controllers
echo -e "${YELLOW}Testing Authentication Controllers...${NC}"
go test -v ./unit_tests/ -run "Test.*Auth.*" -timeout 30s
auth_result=$?

echo ""

# Test 2: Data Management Controllers
echo -e "${YELLOW}Testing Data Management Controllers...${NC}"
go test -v ./unit_tests/ -run "Test.*Management|Test.*Data.*" -timeout 30s
data_result=$?

echo ""

# Test 3: Schedule and Conflict Controllers
echo -e "${YELLOW}Testing Schedule and Conflict Controllers...${NC}"
go test -v ./unit_tests/ -run "Test.*Schedule|Test.*Conflict|Test.*Crosslisting|Test.*API" -timeout 30s
schedule_result=$?

echo ""

# Test 4: Business Logic Validation
echo -e "${YELLOW}Testing Business Logic and Validation...${NC}"
go test -v ./unit_tests/ -run "Test.*Validation|Test.*Logic" -timeout 30s
logic_result=$?

echo ""

# Test 5: Session Management
echo -e "${YELLOW}Testing Session Management...${NC}"
go test -v ./unit_tests/ -run "Test.*Session" -timeout 30s
session_result=$?

echo ""

# Run all tests together for coverage
echo -e "${YELLOW}Running complete test suite with coverage...${NC}"
go test -v -cover ./unit_tests/ -timeout 60s -coverprofile=coverage.out
coverage_result=$?

echo ""
echo "========================================"
echo "Test Results Summary:"
echo "========================================"

# Check results and display summary
if [ $auth_result -eq 0 ]; then
    echo -e "Authentication Controllers: ${GREEN}PASSED${NC}"
else
    echo -e "Authentication Controllers: ${RED}FAILED${NC}"
fi

if [ $data_result -eq 0 ]; then
    echo -e "Data Management Controllers: ${GREEN}PASSED${NC}"
else
    echo -e "Data Management Controllers: ${RED}FAILED${NC}"
fi

if [ $schedule_result -eq 0 ]; then
    echo -e "Schedule Controllers: ${GREEN}PASSED${NC}"
else
    echo -e "Schedule Controllers: ${RED}FAILED${NC}"
fi

if [ $logic_result -eq 0 ]; then
    echo -e "Business Logic Validation: ${GREEN}PASSED${NC}"
else
    echo -e "Business Logic Validation: ${RED}FAILED${NC}"
fi

if [ $session_result -eq 0 ]; then
    echo -e "Session Management: ${GREEN}PASSED${NC}"
else
    echo -e "Session Management: ${RED}FAILED${NC}"
fi

if [ $coverage_result -eq 0 ]; then
    echo -e "Coverage Report: ${GREEN}GENERATED${NC}"
    echo ""
    echo "Coverage details:"
    go tool cover -func=coverage.out | tail -n 1
    echo ""
    echo "To view detailed coverage report, run:"
    echo "go tool cover -html=coverage.out -o coverage.html"
    echo "open coverage.html"
else
    echo -e "Coverage Report: ${RED}FAILED${NC}"
fi

echo ""

# Calculate overall result
total_failures=$((auth_result + data_result + schedule_result + logic_result + session_result))

if [ $total_failures -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed successfully!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please review the output above.${NC}"
    exit 1
fi
