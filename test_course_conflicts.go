package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// Test function for extractNumericCourseNumber
func extractNumericCourseNumber(courseNum string) int {
	// Remove any trailing letters (like H for honors, W for writing intensive, etc.)
	re := regexp.MustCompile(`^(\d+)`)
	matches := re.FindStringSubmatch(courseNum)

	if len(matches) < 2 {
		return -1 // Invalid course number format
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1
	}

	return num
}

// Test function for isInSameCourseRange
func isInSameCourseRange(courseNum1, courseNum2 string) bool {
	num1 := extractNumericCourseNumber(courseNum1)
	num2 := extractNumericCourseNumber(courseNum2)

	if num1 == -1 || num2 == -1 {
		return false // If we can't parse the course numbers, assume no conflict
	}

	// Define the ranges
	ranges := [][2]int{
		{1000, 1999},
		{2000, 2999},
		{3000, 3999},
		{5000, 5999},
		{6000, 6999},
	}

	// Check if both course numbers fall in the same range
	for _, r := range ranges {
		if num1 >= r[0] && num1 <= r[1] && num2 >= r[0] && num2 <= r[1] {
			return true
		}
	}

	return false
}

func main() {
	// Test cases for course number extraction
	testCases := []struct {
		input    string
		expected int
	}{
		{"2150", 2150},
		{"2150H", 2150},
		{"1050W", 1050},
		{"5000", 5000},
		{"6999", 6999},
		{"abc", -1},
		{"", -1},
	}

	fmt.Println("Testing course number extraction:")
	for _, tc := range testCases {
		result := extractNumericCourseNumber(tc.input)
		status := "✓"
		if result != tc.expected {
			status = "✗"
		}
		fmt.Printf("%s Input: '%s' -> Expected: %d, Got: %d\n", status, tc.input, tc.expected, result)
	}

	// Test cases for course range conflicts
	conflictTests := []struct {
		course1        string
		course2        string
		shouldConflict bool
		description    string
	}{
		{"2150", "2200", true, "Both in 2000-2999 range"},
		{"1050", "1999", true, "Both in 1000-1999 range"},
		{"5000", "5999", true, "Both in 5000-5999 range"},
		{"6000", "6500", true, "Both in 6000-6999 range"},
		{"3000", "3999", true, "Both in 3000-3999 range"},
		{"2150", "3150", false, "Different ranges (2000s vs 3000s)"},
		{"1050", "2050", false, "Different ranges (1000s vs 2000s)"},
		{"2150H", "2200W", true, "Both in 2000s with letters"},
		{"4000", "4500", false, "4000s range not covered"},
		{"900", "950", false, "Below 1000"},
		{"7000", "7500", false, "Above 6999"},
	}

	fmt.Println("\nTesting course range conflicts:")
	for _, tc := range conflictTests {
		result := isInSameCourseRange(tc.course1, tc.course2)
		status := "✓"
		if result != tc.shouldConflict {
			status = "✗"
		}
		fmt.Printf("%s %s vs %s -> Expected: %t, Got: %t (%s)\n",
			status, tc.course1, tc.course2, tc.shouldConflict, result, tc.description)
	}
}
