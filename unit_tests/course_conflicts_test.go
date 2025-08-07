package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test data structures for course conflicts
type CourseConflictPair struct {
	Course1 CourseConflictDetail
	Course2 CourseConflictDetail
	Type    string
}

type CourseConflictDetail struct {
	ID           int
	CRN          int
	Section      string
	ScheduleID   int
	Prefix       string
	CourseNumber string
	Title        string
	InstructorID int
	TimeSlotID   int
	RoomID       int
	Mode         string
	Lab          bool
	TimeSlot     *CourseConflictTimeSlot
}

type CourseConflictTimeSlot struct {
	ID        int
	StartTime string
	EndTime   string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Days      string
}

type CourseConflictPrerequisite struct {
	ID                int
	PredPrefixID      int
	PredCourseNum     string
	SuccPrefixID      int
	SuccCourseNum     string
	PredecessorPrefix string
	PredecessorNumber string
	SuccessorPrefix   string
	SuccessorNumber   string
}

// MockCourseConflictScheduler implements the scheduler interface for testing course conflicts
type MockCourseConflictScheduler struct {
	mock.Mock
	prerequisites []CourseConflictPrerequisite
	crosslistings map[string]bool
}

// Mock methods needed for course conflict detection
func (m *MockCourseConflictScheduler) GetAllPrerequisites() ([]CourseConflictPrerequisite, error) {
	args := m.Called()
	return args.Get(0).([]CourseConflictPrerequisite), args.Error(1)
}

func (m *MockCourseConflictScheduler) AreCoursesCrosslisted(crn1, crn2 int) (bool, error) {
	args := m.Called(crn1, crn2)
	return args.Bool(0), args.Error(1)
}

func (m *MockCourseConflictScheduler) timeSlotsOverlap(ts1, ts2 *CourseConflictTimeSlot) bool {
	if ts1 == nil || ts2 == nil {
		return false
	}

	// Check if they share any common days
	daysOverlap := (ts1.Monday && ts2.Monday) ||
		(ts1.Tuesday && ts2.Tuesday) ||
		(ts1.Wednesday && ts2.Wednesday) ||
		(ts1.Thursday && ts2.Thursday) ||
		(ts1.Friday && ts2.Friday)

	if !daysOverlap {
		return false
	}

	// Check if times overlap - simple string comparison for HH:MM format
	return ts1.StartTime < ts2.EndTime && ts2.StartTime < ts1.EndTime
}

// extractNumericCourseNumber - copy of the actual function for testing
func (m *MockCourseConflictScheduler) extractNumericCourseNumber(courseNum string) int {
	// Simple implementation for testing - extract first sequence of digits
	var result string
	for _, char := range courseNum {
		if char >= '0' && char <= '9' {
			result += string(char)
		} else if result != "" {
			break // Stop at first non-digit after finding digits
		}
	}

	if result == "" {
		return -1
	}

	// Convert to int
	num := 0
	for _, char := range result {
		num = num*10 + int(char-'0')
	}
	return num
}

// isInSameCourseRange - copy of the actual function for testing
func (m *MockCourseConflictScheduler) isInSameCourseRange(courseNum1, courseNum2 string) bool {
	num1 := m.extractNumericCourseNumber(courseNum1)
	num2 := m.extractNumericCourseNumber(courseNum2)

	if num1 == -1 || num2 == -1 {
		return false
	}

	ranges := [][2]int{
		{1000, 1999},
		{2000, 2999},
		{3000, 3999},
		{5000, 5999},
		{6000, 6999},
	}

	for _, r := range ranges {
		if num1 >= r[0] && num1 <= r[1] && num2 >= r[0] && num2 <= r[1] {
			return true
		}
	}

	return false
}

// isCourseConflictException - copy of the actual function for testing
func (m *MockCourseConflictScheduler) isCourseConflictException(course1, course2 CourseConflictDetail) (bool, error) {
	// Check if courses are crosslisted
	crosslisted, err := m.AreCoursesCrosslisted(course1.CRN, course2.CRN)
	if err != nil {
		return false, fmt.Errorf("error checking crosslisting: %v", err)
	}
	if crosslisted {
		return true, nil
	}

	// Check if courses are on the same prerequisite chain
	onSameChain, err := m.areCoursesOnSamePrerequisiteChain(course1.Prefix, course1.CourseNumber, course2.Prefix, course2.CourseNumber)
	if err != nil {
		return false, fmt.Errorf("error checking prerequisite chain: %v", err)
	}
	if onSameChain {
		return true, nil
	}

	return false, nil
}

// areCoursesOnSamePrerequisiteChain - copy of the actual function for testing
func (m *MockCourseConflictScheduler) areCoursesOnSamePrerequisiteChain(prefix1, courseNum1, prefix2, courseNum2 string) (bool, error) {
	prerequisites, err := m.GetAllPrerequisites()
	if err != nil {
		return false, fmt.Errorf("failed to get prerequisites: %v", err)
	}

	// Build a graph of prerequisite relationships
	prereqGraph := make(map[string][]string) // course -> list of prerequisite courses
	succGraph := make(map[string][]string)   // course -> list of successor courses

	for _, prereq := range prerequisites {
		predCourse := prereq.PredecessorPrefix + " " + prereq.PredecessorNumber
		succCourse := prereq.SuccessorPrefix + " " + prereq.SuccessorNumber

		prereqGraph[succCourse] = append(prereqGraph[succCourse], predCourse)
		succGraph[predCourse] = append(succGraph[predCourse], succCourse)
	}

	course1Key := prefix1 + " " + courseNum1
	course2Key := prefix2 + " " + courseNum2

	// Check if course1 is a prerequisite for course2 (directly or indirectly)
	if m.isPrerequisiteOf(course1Key, course2Key, prereqGraph, make(map[string]bool)) {
		return true, nil
	}

	// Check if course2 is a prerequisite for course1 (directly or indirectly)
	if m.isPrerequisiteOf(course2Key, course1Key, prereqGraph, make(map[string]bool)) {
		return true, nil
	}

	return false, nil
}

// isPrerequisiteOf - copy of the actual function for testing
func (m *MockCourseConflictScheduler) isPrerequisiteOf(course1, course2 string, prereqGraph map[string][]string, visited map[string]bool) bool {
	if visited[course2] {
		return false // Avoid infinite loops
	}
	visited[course2] = true

	prerequisites, exists := prereqGraph[course2]
	if !exists {
		return false
	}

	// Check direct prerequisite
	for _, prereq := range prerequisites {
		if prereq == course1 {
			return true
		}
	}

	// Check indirect prerequisite (recursive)
	for _, prereq := range prerequisites {
		if m.isPrerequisiteOf(course1, prereq, prereqGraph, visited) {
			return true
		}
	}

	return false
}

// detectCourseConflicts - copy of the actual function for testing
func (m *MockCourseConflictScheduler) detectCourseConflicts(courses1, courses2 []CourseConflictDetail) ([]CourseConflictPair, error) {
	var courseConflicts []CourseConflictPair

	// Create a map to track unique courses by CRN to avoid duplicates
	courseMap := make(map[int]CourseConflictDetail)

	// Add all courses to the map, preferring courses from courses1 if duplicates exist
	for _, course := range courses1 {
		courseMap[course.CRN] = course
	}
	for _, course := range courses2 {
		if _, exists := courseMap[course.CRN]; !exists {
			courseMap[course.CRN] = course
		}
	}

	// Convert map back to slice for processing
	allCourses := make([]CourseConflictDetail, 0, len(courseMap))
	for _, course := range courseMap {
		allCourses = append(allCourses, course)
	}

	// Check all unique course pairs for course conflicts
	for i, course1 := range allCourses {
		for j, course2 := range allCourses {
			if i >= j { // Avoid checking the same pair twice and avoid self-comparison
				continue
			}

			// Only check courses with the same prefix
			if course1.Prefix != course2.Prefix {
				continue
			}

			// Check if time slots overlap
			if !m.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
				continue
			}

			// Mode exception: Courses with same prefix and course number but different modes don't conflict
			if course1.Prefix == course2.Prefix && course1.CourseNumber == course2.CourseNumber &&
				course1.Mode != course2.Mode {
				continue
			}

			// Lab-specific logic: Labs don't conflict with any other courses (including other labs)
			// EXCEPT: Labs may not be offered at the same time as the same course number that is not a lab
			if course1.Lab || course2.Lab {
				// If one is a lab and the other is not a lab AND they have the same course number, it's a conflict
				if course1.Lab != course2.Lab && course1.CourseNumber == course2.CourseNumber {
					conflictPair := CourseConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "course",
					}
					courseConflicts = append(courseConflicts, conflictPair)
				}
				// If both are labs, or they have different course numbers, no conflict - skip to next pair
				continue
			}

			// For non-lab courses, check if courses are in the same course number range and would conflict
			if m.isInSameCourseRange(course1.CourseNumber, course2.CourseNumber) {
				// Check for exceptions: crosslisted courses or prerequisite chain
				isException, err := m.isCourseConflictException(course1, course2)
				if err != nil {
					// In tests, we'll log errors but continue
					continue
				}

				if !isException {
					conflictPair := CourseConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "course",
					}
					courseConflicts = append(courseConflicts, conflictPair)
				}
			}
		}
	}

	return courseConflicts, nil
}

// Test cases for extractNumericCourseNumber
func TestExtractNumericCourseNumber(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	testCases := []struct {
		input    string
		expected int
		name     string
	}{
		{"2150", 2150, "Basic course number"},
		{"2150H", 2150, "Honors course"},
		{"1050W", 1050, "Writing intensive course"},
		{"5000", 5000, "Graduate course"},
		{"6999", 6999, "High graduate course"},
		{"abc", -1, "Invalid non-numeric"},
		{"", -1, "Empty string"},
		{"12abc34", 12, "Mixed with non-numeric in middle"},
		{"CS2150", 2150, "With prefix"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockScheduler.extractNumericCourseNumber(tc.input)
			assert.Equal(t, tc.expected, result, "Expected %d for input '%s', got %d", tc.expected, tc.input, result)
		})
	}
}

// Test cases for isInSameCourseRange
func TestIsInSameCourseRange(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	testCases := []struct {
		course1        string
		course2        string
		shouldConflict bool
		name           string
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockScheduler.isInSameCourseRange(tc.course1, tc.course2)
			assert.Equal(t, tc.shouldConflict, result,
				"Expected %t for %s vs %s, got %t", tc.shouldConflict, tc.course1, tc.course2, result)
		})
	}
}

// Test cases for timeSlotsOverlap
func TestTimeSlotsOverlap(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	testCases := []struct {
		slot1         *CourseConflictTimeSlot
		slot2         *CourseConflictTimeSlot
		shouldOverlap bool
		name          string
	}{
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true, Wednesday: true, Friday: true},
			&CourseConflictTimeSlot{StartTime: "10:30", EndTime: "11:30", Monday: true, Wednesday: true, Friday: true},
			true,
			"Overlapping time on same days",
		},
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true, Wednesday: true, Friday: true},
			&CourseConflictTimeSlot{StartTime: "11:00", EndTime: "12:00", Monday: true, Wednesday: true, Friday: true},
			false,
			"Adjacent time slots (no overlap)",
		},
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true, Wednesday: true, Friday: true},
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Tuesday: true, Thursday: true},
			false,
			"Same time, different days",
		},
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true, Tuesday: true},
			&CourseConflictTimeSlot{StartTime: "10:30", EndTime: "11:30", Monday: true, Wednesday: true},
			true,
			"Overlapping time with one shared day",
		},
		{
			nil,
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true},
			false,
			"One nil timeslot",
		},
		{
			nil,
			nil,
			false,
			"Both nil timeslots",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockScheduler.timeSlotsOverlap(tc.slot1, tc.slot2)
			assert.Equal(t, tc.shouldOverlap, result,
				"Expected %t for timeslot overlap, got %t", tc.shouldOverlap, result)
		})
	}
}

// Test cases for prerequisite chain detection
func TestAreCoursesOnSamePrerequisiteChain(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup prerequisite chain: MATH 1050 -> MATH 2150 -> MATH 3150
	prerequisites := []CourseConflictPrerequisite{
		{
			ID: 1, PredPrefixID: 1, PredCourseNum: "1050", SuccPrefixID: 1, SuccCourseNum: "2150",
			PredecessorPrefix: "MATH", PredecessorNumber: "1050",
			SuccessorPrefix: "MATH", SuccessorNumber: "2150",
		},
		{
			ID: 2, PredPrefixID: 1, PredCourseNum: "2150", SuccPrefixID: 1, SuccCourseNum: "3150",
			PredecessorPrefix: "MATH", PredecessorNumber: "2150",
			SuccessorPrefix: "MATH", SuccessorNumber: "3150",
		},
		{
			ID: 3, PredPrefixID: 2, PredCourseNum: "1000", SuccPrefixID: 2, SuccCourseNum: "2000",
			PredecessorPrefix: "CS", PredecessorNumber: "1000",
			SuccessorPrefix: "CS", SuccessorNumber: "2000",
		},
	}

	mockScheduler.On("GetAllPrerequisites").Return(prerequisites, nil)

	testCases := []struct {
		prefix1       string
		courseNum1    string
		prefix2       string
		courseNum2    string
		expectedChain bool
		name          string
	}{
		{"MATH", "1050", "MATH", "2150", true, "Direct prerequisite relationship"},
		{"MATH", "1050", "MATH", "3150", true, "Indirect prerequisite relationship"},
		{"MATH", "2150", "MATH", "3150", true, "Direct prerequisite relationship (reverse order)"},
		{"MATH", "1050", "CS", "1000", false, "Different prefixes"},
		{"MATH", "1050", "MATH", "1060", false, "No prerequisite relationship"},
		{"CS", "1000", "CS", "2000", true, "Different prefix chain"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mockScheduler.areCoursesOnSamePrerequisiteChain(tc.prefix1, tc.courseNum1, tc.prefix2, tc.courseNum2)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedChain, result,
				"Expected %t for %s %s and %s %s on same chain, got %t",
				tc.expectedChain, tc.prefix1, tc.courseNum1, tc.prefix2, tc.courseNum2, result)
		})
	}
}

// Test course conflict detection with no exceptions
func TestDetectCourseConflicts_NoExceptions(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 1, "Expected 1 conflict")
	assert.Equal(t, "course", conflicts[0].Type)
	assert.Equal(t, 12345, conflicts[0].Course1.CRN)
	assert.Equal(t, 12346, conflicts[0].Course2.CRN)
}

// Test course conflict detection with crosslisting exception
func TestDetectCourseConflicts_CrosslistedExecution(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls - courses are crosslisted, so no conflict expected
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Expected no conflicts due to crosslisting exception")
}

// Test course conflict detection with prerequisite chain exception
func TestDetectCourseConflicts_PrerequisiteChainException(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup prerequisite chain: CS 2150 -> CS 2200
	prerequisites := []CourseConflictPrerequisite{
		{
			ID: 1, PredPrefixID: 1, PredCourseNum: "2150", SuccPrefixID: 1, SuccCourseNum: "2200",
			PredecessorPrefix: "CS", PredecessorNumber: "2150",
			SuccessorPrefix: "CS", SuccessorNumber: "2200",
		},
	}

	mockScheduler.On("GetAllPrerequisites").Return(prerequisites, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Expected no conflicts due to prerequisite chain exception")
}

// Test no conflicts when courses are in different ranges
func TestDetectCourseConflicts_DifferentRanges(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "1050", Title: "Intro Programming", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Expected no conflicts - different course number ranges")
}

// Test no conflicts when courses have different prefixes
func TestDetectCourseConflicts_DifferentPrefixes(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "MATH", CourseNumber: "2150", Title: "Calculus II", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Expected no conflicts - different prefixes")
}

// Test no conflicts when time slots don't overlap
func TestDetectCourseConflicts_NoTimeOverlap(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	timeSlot1 := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	timeSlot2 := &CourseConflictTimeSlot{
		StartTime: "11:00", EndTime: "12:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot1},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: timeSlot2},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Expected no conflicts - no time overlap")
}

// Test multiple conflicts detected
func TestDetectCourseConflicts_MultipleConflicts(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12347, Prefix: "CS", CourseNumber: "2300", Title: "Software Engineering", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 3, "Expected 3 conflicts (all pairs)")

	// Verify all conflicts are course conflicts
	for _, conflict := range conflicts {
		assert.Equal(t, "course", conflict.Type)
	}
}

// Test edge cases and error scenarios
func TestDetectCourseConflicts_EdgeCases(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	t.Run("Empty course lists", func(t *testing.T) {
		conflicts, err := mockScheduler.detectCourseConflicts([]CourseConflictDetail{}, []CourseConflictDetail{})
		assert.NoError(t, err)
		assert.Len(t, conflicts, 0)
	})

	t.Run("Single course list", func(t *testing.T) {
		timeSlot := &CourseConflictTimeSlot{
			StartTime: "10:00", EndTime: "11:00",
			Monday: true, Wednesday: true, Friday: true,
		}

		courses := []CourseConflictDetail{
			{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
		}

		conflicts, err := mockScheduler.detectCourseConflicts(courses, []CourseConflictDetail{})
		assert.NoError(t, err)
		assert.Len(t, conflicts, 0)
	})

	t.Run("Duplicate CRNs", func(t *testing.T) {
		mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
		mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

		timeSlot := &CourseConflictTimeSlot{
			StartTime: "10:00", EndTime: "11:00",
			Monday: true, Wednesday: true, Friday: true,
		}

		// Same CRN in both lists - should be deduplicated
		course := CourseConflictDetail{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot}

		conflicts, err := mockScheduler.detectCourseConflicts([]CourseConflictDetail{course}, []CourseConflictDetail{course})
		assert.NoError(t, err)
		assert.Len(t, conflicts, 0, "Duplicate CRNs should be deduplicated")
	})

	t.Run("Courses with nil time slots", func(t *testing.T) {
		courses1 := []CourseConflictDetail{
			{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: nil},
		}

		courses2 := []CourseConflictDetail{
			{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: nil},
		}

		conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)
		assert.NoError(t, err)
		assert.Len(t, conflicts, 0, "Courses with nil time slots should not conflict")
	})
}

// Test boundary conditions for course number ranges
func TestCourseNumberRangeBoundaries(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	testCases := []struct {
		course1        string
		course2        string
		shouldConflict bool
		name           string
	}{
		{"1000", "1999", true, "Boundary values for 1000s range"},
		{"2000", "2999", true, "Boundary values for 2000s range"},
		{"3000", "3999", true, "Boundary values for 3000s range"},
		{"5000", "5999", true, "Boundary values for 5000s range"},
		{"6000", "6999", true, "Boundary values for 6000s range"},
		{"999", "1000", false, "Boundary crossing below 1000s"},
		{"1999", "2000", false, "Boundary crossing 1000s to 2000s"},
		{"2999", "3000", false, "Boundary crossing 2000s to 3000s"},
		{"3999", "4000", false, "Boundary crossing 3000s to 4000s (gap)"},
		{"4999", "5000", false, "Boundary crossing 4000s to 5000s"},
		{"5999", "6000", false, "Boundary crossing 5000s to 6000s"},
		{"6999", "7000", false, "Boundary crossing above 6000s"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockScheduler.isInSameCourseRange(tc.course1, tc.course2)
			assert.Equal(t, tc.shouldConflict, result,
				"Expected %t for %s vs %s, got %t", tc.shouldConflict, tc.course1, tc.course2, result)
		})
	}
}

// Test complex prerequisite chains
func TestComplexPrerequisiteChains(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup complex prerequisite chain: MATH 1050 -> MATH 2150 -> MATH 3150 -> MATH 5150
	//                                  CS 1000 -> CS 2000 -> CS 3000
	prerequisites := []CourseConflictPrerequisite{
		{ID: 1, PredecessorPrefix: "MATH", PredecessorNumber: "1050", SuccessorPrefix: "MATH", SuccessorNumber: "2150"},
		{ID: 2, PredecessorPrefix: "MATH", PredecessorNumber: "2150", SuccessorPrefix: "MATH", SuccessorNumber: "3150"},
		{ID: 3, PredecessorPrefix: "MATH", PredecessorNumber: "3150", SuccessorPrefix: "MATH", SuccessorNumber: "5150"},
		{ID: 4, PredecessorPrefix: "CS", PredecessorNumber: "1000", SuccessorPrefix: "CS", SuccessorNumber: "2000"},
		{ID: 5, PredecessorPrefix: "CS", PredecessorNumber: "2000", SuccessorPrefix: "CS", SuccessorNumber: "3000"},
	}

	mockScheduler.On("GetAllPrerequisites").Return(prerequisites, nil)

	testCases := []struct {
		prefix1       string
		courseNum1    string
		prefix2       string
		courseNum2    string
		expectedChain bool
		name          string
	}{
		{"MATH", "1050", "MATH", "5150", true, "Long prerequisite chain (3 steps)"},
		{"MATH", "5150", "MATH", "1050", true, "Reverse long prerequisite chain"},
		{"CS", "1000", "CS", "3000", true, "Two-step prerequisite chain"},
		{"MATH", "1050", "CS", "1000", false, "Different prefix chains"},
		{"MATH", "2150", "MATH", "5150", true, "Partial chain overlap"},
		{"MATH", "1050", "MATH", "1050", false, "Same course (self-reference)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mockScheduler.areCoursesOnSamePrerequisiteChain(tc.prefix1, tc.courseNum1, tc.prefix2, tc.courseNum2)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedChain, result,
				"Expected %t for %s %s and %s %s on same chain, got %t",
				tc.expectedChain, tc.prefix1, tc.courseNum1, tc.prefix2, tc.courseNum2, result)
		})
	}
}

// Test time slot edge cases
func TestTimeSlotEdgeCases(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	testCases := []struct {
		slot1         *CourseConflictTimeSlot
		slot2         *CourseConflictTimeSlot
		shouldOverlap bool
		name          string
	}{
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "10:50", Monday: true},
			&CourseConflictTimeSlot{StartTime: "10:50", EndTime: "11:40", Monday: true},
			false,
			"Exact time boundary (no overlap)",
		},
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "10:51", Monday: true},
			&CourseConflictTimeSlot{StartTime: "10:50", EndTime: "11:40", Monday: true},
			true,
			"One minute overlap",
		},
		{
			&CourseConflictTimeSlot{StartTime: "08:00", EndTime: "23:59", Monday: true},
			&CourseConflictTimeSlot{StartTime: "12:00", EndTime: "13:00", Monday: true},
			true,
			"Long time slot containing shorter one",
		},
		{
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true, Tuesday: true},
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Tuesday: true, Wednesday: true},
			true,
			"Overlapping on one shared day",
		},
		{
			&CourseConflictTimeSlot{StartTime: "", EndTime: "", Monday: true},
			&CourseConflictTimeSlot{StartTime: "10:00", EndTime: "11:00", Monday: true},
			false,
			"Empty time strings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mockScheduler.timeSlotsOverlap(tc.slot1, tc.slot2)
			assert.Equal(t, tc.shouldOverlap, result,
				"Expected %t for timeslot overlap, got %t", tc.shouldOverlap, result)
		})
	}
}

// Benchmark test for performance with large course lists
func BenchmarkDetectCourseConflicts(b *testing.B) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	// Create large course lists
	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	var courses1, courses2 []CourseConflictDetail
	for i := 0; i < 100; i++ {
		courses1 = append(courses1, CourseConflictDetail{
			CRN: 10000 + i, Prefix: "CS", CourseNumber: "2150",
			Title: fmt.Sprintf("Course %d", i), Lab: false, TimeSlot: timeSlot,
		})
		courses2 = append(courses2, CourseConflictDetail{
			CRN: 20000 + i, Prefix: "CS", CourseNumber: "2200",
			Title: fmt.Sprintf("Course %d", i+100), Lab: false, TimeSlot: timeSlot,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockScheduler.detectCourseConflicts(courses1, courses2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Lab-specific conflict detection tests
func TestDetectCourseConflicts_LabNoConflictWithOtherCourses(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test lab vs non-lab with same prefix, different course numbers, same time
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Lab: true, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Labs should not conflict with other courses")
}

func TestDetectCourseConflicts_LabNoConflictWithOtherLabs(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test lab vs lab with same prefix, different course numbers, same time
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Lab: true, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2200", Title: "Computer Organization Lab", Lab: true, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Labs should not conflict with other labs")
}

func TestDetectCourseConflicts_LabConflictWithSameCourseNumber(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test lab vs non-lab with same prefix, same course number, same time - should conflict
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Lab: true, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 1, "Lab should conflict with non-lab of same course number")
	assert.Equal(t, "course", conflicts[0].Type)
}

func TestDetectCourseConflicts_LabNoConflictDifferentTime(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot1 := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	timeSlot2 := &CourseConflictTimeSlot{
		StartTime: "14:00", EndTime: "15:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test lab vs non-lab with same prefix, same course number, different times - should not conflict
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Lab: true, TimeSlot: timeSlot1},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Lab: false, TimeSlot: timeSlot2},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Lab should not conflict with non-lab when at different times")
}

func TestDetectCourseConflicts_LabsWithDifferentCourseNumbers(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test lab vs non-lab with same prefix, different course numbers, same time - should not conflict
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Lab: true, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "3200", Title: "Algorithms", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Lab should not conflict with non-lab of different course number")
}

// Mode exception test
func TestDetectCourseConflicts_SameCourseNumberDifferentModes(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test courses with same prefix and course number but different modes - should not conflict
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ONL", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Courses with same prefix and course number but different modes should not conflict")
}

func TestDetectCourseConflicts_SameCourseNumberSameMode(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test courses with same prefix, course number, and mode - should conflict
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)

	assert.NoError(t, err)
	assert.Len(t, conflicts, 1, "Courses with same prefix, course number, and mode should conflict")
	assert.Equal(t, "course", conflicts[0].Type)
}

func TestDetectCourseConflicts_ModeAndLabCombinations(t *testing.T) {
	mockScheduler := &MockCourseConflictScheduler{}

	// Setup mock calls
	mockScheduler.On("GetAllPrerequisites").Return([]CourseConflictPrerequisite{}, nil)
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(false, nil)

	timeSlot := &CourseConflictTimeSlot{
		StartTime: "10:00", EndTime: "11:00",
		Monday: true, Wednesday: true, Friday: true,
	}

	// Test 1: Lab and non-lab with same course number but different modes - should not conflict (mode exception takes precedence)
	courses1 := []CourseConflictDetail{
		{CRN: 12345, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Mode: "ITP", Lab: true, TimeSlot: timeSlot},
	}

	courses2 := []CourseConflictDetail{
		{CRN: 12346, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ONL", Lab: false, TimeSlot: timeSlot},
	}

	conflicts, err := mockScheduler.detectCourseConflicts(courses1, courses2)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 0, "Lab and non-lab with same course number but different modes should not conflict")

	// Test 2: Lab and non-lab with same course number and same mode - should conflict (lab exception applies)
	courses3 := []CourseConflictDetail{
		{CRN: 12347, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures Lab", Mode: "ITP", Lab: true, TimeSlot: timeSlot},
	}

	courses4 := []CourseConflictDetail{
		{CRN: 12348, Prefix: "CS", CourseNumber: "2150", Title: "Data Structures", Mode: "ITP", Lab: false, TimeSlot: timeSlot},
	}

	conflicts2, err2 := mockScheduler.detectCourseConflicts(courses3, courses4)
	assert.NoError(t, err2)
	assert.Len(t, conflicts2, 1, "Lab and non-lab with same course number and same mode should conflict")
	assert.Equal(t, "course", conflicts2[0].Type)
}
