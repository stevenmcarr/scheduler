#!/bin/bash

# WMU Course Scheduler - Removed Courses and Crosslisting Test Runner
# This script specifically tests the new logic for "Removed" courses and cross-listed courses

echo "================================================================"
echo "WMU Course Scheduler - Removed Courses & Crosslisting Tests"
echo "================================================================"
echo ""

# Change to the project root directory
cd "$(dirname "$0")/.."

# Set Go environment
export CGO_ENABLED=1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing new logic for 'Removed' courses and cross-listed courses...${NC}"
echo ""

# Test 1: Removed Courses Logic (Simple)
echo -e "${YELLOW}🧪 Testing Removed Courses Logic (Simple Implementation)...${NC}"
echo -e "${CYAN}   - Instructor conflicts with removed courses${NC}"
echo -e "${CYAN}   - Room conflicts with removed courses${NC}"
echo -e "${CYAN}   - Crosslisting conflicts with removed courses${NC}"
echo -e "${CYAN}   - Course conflicts with removed courses${NC}"
echo ""
go test -v ./unit_tests/ -run "TestSimple.*" -timeout 30s
simple_result=$?

echo ""
echo "================================================================"

# Test 2: Removed Courses Logic (Original Mock-based)
echo -e "${YELLOW}🧪 Testing Removed Courses Logic (Mock-based)...${NC}"
echo -e "${CYAN}   - Comprehensive mock-based conflict detection${NC}"
echo ""
go test -v ./unit_tests/ -run "TestRemovedCourses.*" -timeout 30s
removed_result=$?

echo ""
echo "================================================================"

# Test 3: Crosslisted Courses Logic
echo -e "${YELLOW}🔗 Testing Crosslisted Courses Logic...${NC}"
echo -e "${CYAN}   - Same time slot crosslisted courses (should not conflict)${NC}"
echo -e "${CYAN}   - Different time slot crosslisted courses (should conflict)${NC}"
echo -e "${CYAN}   - AO mode crosslisted courses (should not conflict)${NC}"
echo ""
go test -v ./unit_tests/ -run "TestCrosslistedCourses.*" -timeout 30s
crosslisted_result=$?

echo ""
echo "================================================================"

# Test 4: Integration Tests
echo -e "${YELLOW}🧩 Testing Integration Scenarios...${NC}"
echo -e "${CYAN}   - Multiple removed courses in complex scenarios${NC}"
echo -e "${CYAN}   - Crosslisted courses with different modes${NC}"
echo -e "${CYAN}   - Removed courses in crosslisting groups${NC}"
echo -e "${CYAN}   - All conflict types with removed status verification${NC}"
echo -e "${CYAN}   - Large scale performance testing${NC}"
echo ""
go test -v ./unit_tests/ -run "TestIntegration.*" -timeout 45s
integration_result=$?

echo ""
echo "================================================================"

# Test 5: Edge Cases
echo -e "${YELLOW}🎯 Testing Edge Cases...${NC}"
echo -e "${CYAN}   - Both courses removed${NC}"
echo -e "${CYAN}   - Mixed status combinations${NC}"
echo -e "${CYAN}   - Normal courses still working correctly${NC}"
echo ""
go test -v ./unit_tests/ -run "TestBothCoursesRemoved|TestMixedStatusCombinations|TestNormalCourses.*" -timeout 30s
edge_result=$?

echo ""
echo "================================================================"

# Run specific test files with coverage
echo -e "${YELLOW}📊 Running all new tests with coverage...${NC}"
go test -v -cover ./unit_tests/ -run "TestSimple|TestRemovedCourses|TestCrosslistedCourses|TestIntegration|TestBothCoursesRemoved|TestMixedStatusCombinations|TestNormalCourses" -timeout 90s -coverprofile=removed_crosslisting_coverage.out
coverage_result=$?

echo ""
echo "================================================================"
echo "Test Results Summary:"
echo "================================================================"

# Check results and display summary
if [ $simple_result -eq 0 ]; then
    echo -e "Simple Removed Courses Tests: ${GREEN}✅ PASSED${NC}"
else
    echo -e "Simple Removed Courses Tests: ${RED}❌ FAILED${NC}"
fi

if [ $removed_result -eq 0 ]; then
    echo -e "Mock-based Removed Courses Tests: ${GREEN}✅ PASSED${NC}"
else
    echo -e "Mock-based Removed Courses Tests: ${RED}❌ FAILED${NC}"
fi

if [ $crosslisted_result -eq 0 ]; then
    echo -e "Crosslisted Courses Logic: ${GREEN}✅ PASSED${NC}"
else
    echo -e "Crosslisted Courses Logic: ${RED}❌ FAILED${NC}"
fi

if [ $integration_result -eq 0 ]; then
    echo -e "Integration Scenarios: ${GREEN}✅ PASSED${NC}"
else
    echo -e "Integration Scenarios: ${RED}❌ FAILED${NC}"
fi

if [ $edge_result -eq 0 ]; then
    echo -e "Edge Cases: ${GREEN}✅ PASSED${NC}"
else
    echo -e "Edge Cases: ${RED}❌ FAILED${NC}"
fi

if [ $coverage_result -eq 0 ]; then
    echo -e "Coverage Report: ${GREEN}✅ GENERATED${NC}"
    echo ""
    echo -e "${CYAN}Coverage details for new tests:${NC}"
    go tool cover -func=removed_crosslisting_coverage.out | grep -E "(TestRemovedCourses|TestCrosslistedCourses|TestIntegration)" | head -10
    echo ""
    echo -e "${BLUE}To view detailed coverage report:${NC}"
    echo -e "${CYAN}go tool cover -html=removed_crosslisting_coverage.out -o removed_crosslisting_coverage.html${NC}"
    echo -e "${CYAN}open removed_crosslisting_coverage.html${NC}"
else
    echo -e "Coverage Report: ${RED}❌ FAILED${NC}"
fi

echo ""

# Calculate overall result
total_failures=$((removed_result + crosslisted_result + integration_result + edge_result))

if [ $total_failures -eq 0 ]; then
    echo -e "${GREEN}🎉 All new tests passed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Key Features Tested:${NC}"
    echo -e "${CYAN}• Removed courses do not create instructor conflicts${NC}"
    echo -e "${CYAN}• Removed courses do not create room conflicts${NC}"
    echo -e "${CYAN}• Removed courses do not create crosslisting conflicts${NC}"
    echo -e "${CYAN}• Removed courses do not create course conflicts${NC}"
    echo -e "${CYAN}• Crosslisted courses with same time slots do not conflict${NC}"
    echo -e "${CYAN}• Crosslisted courses with different time slots do conflict${NC}"
    echo -e "${CYAN}• AO mode courses are exempt from time-based conflicts${NC}"
    echo -e "${CYAN}• Normal course conflicts still work properly${NC}"
    echo -e "${CYAN}• Edge cases and complex scenarios handled correctly${NC}"
    echo ""
    echo -e "${GREEN}The conflict detection system properly implements:${NC}"
    echo -e "${GREEN}1. 'Removed' status courses cannot conflict with any other course${NC}"
    echo -e "${GREEN}2. Cross-listed courses are properly exempted from inappropriate conflicts${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Some new tests failed. Please review the output above.${NC}"
    echo ""
    echo -e "${YELLOW}Failed test categories:${NC}"
    [ $removed_result -ne 0 ] && echo -e "${RED}• Removed Courses Logic${NC}"
    [ $crosslisted_result -ne 0 ] && echo -e "${RED}• Crosslisted Courses Logic${NC}"
    [ $integration_result -ne 0 ] && echo -e "${RED}• Integration Scenarios${NC}"
    [ $edge_result -ne 0 ] && echo -e "${RED}• Edge Cases${NC}"
    echo ""
    exit 1
fi
