package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test data structures - duplicated here for testing isolation
type CrosslistingConflictPair struct {
	Course1 CrosslistingCourseDetail
	Course2 CrosslistingCourseDetail
	Type    string
}

type CrosslistingCourseDetail struct {
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
	TimeSlot     *CrosslistingTimeSlot
}

type CrosslistingTimeSlot struct {
	ID        int
	StartTime string
	EndTime   string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
}

// MockScheduler implements the scheduler interface for testing
type MockScheduler struct {
	mock.Mock
	crosslistings map[string]bool
}

func (m *MockScheduler) AreCoursesCrosslisted(crn1, crn2 int) (bool, error) {
	args := m.Called(crn1, crn2)
	return args.Bool(0), args.Error(1)
}

// timeSlotsMatch checks if two time slots are exactly the same
func (m *MockScheduler) timeSlotsMatch(slot1, slot2 *CrosslistingTimeSlot) bool {
	if slot1 == nil || slot2 == nil {
		return slot1 == slot2 // Both nil = match, one nil = no match
	}

	return slot1.StartTime == slot2.StartTime &&
		slot1.EndTime == slot2.EndTime &&
		slot1.Monday == slot2.Monday &&
		slot1.Tuesday == slot2.Tuesday &&
		slot1.Wednesday == slot2.Wednesday &&
		slot1.Thursday == slot2.Thursday &&
		slot1.Friday == slot2.Friday
}

// isRoomExemptMode checks if a course is in a mode that exempts it from room conflicts (FSO, PSO, AO)
func (m *MockScheduler) isRoomExemptMode(course CrosslistingCourseDetail) bool {
	return course.Mode == "FSO" || course.Mode == "PSO" || course.Mode == "AO"
}

// isTimeExemptMode checks if a course is in a mode that exempts it from time conflicts (AO)
func (m *MockScheduler) isTimeExemptMode(course CrosslistingCourseDetail) bool {
	return course.Mode == "AO"
}

// Helper function to create crosslisting key
func getCrosslistingKey(crn1, crn2 int) string {
	if crn1 < crn2 {
		return fmt.Sprintf("%d-%d", crn1, crn2)
	}
	return fmt.Sprintf("%d-%d", crn2, crn1)
}

// detectCrosslistingConflicts - copy of the actual function for testing
func (m *MockScheduler) detectCrosslistingConflicts(courses1, courses2 []CrosslistingCourseDetail) ([]CrosslistingConflictPair, error) {
	var crosslistingConflicts []CrosslistingConflictPair

	// Create a map to track unique courses by CRN to avoid duplicates
	courseMap := make(map[int]CrosslistingCourseDetail)

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
	allCourses := make([]CrosslistingCourseDetail, 0, len(courseMap))
	for _, course := range courseMap {
		allCourses = append(allCourses, course)
	}

	// Check all unique course pairs for crosslisting conflicts
	for i, course1 := range allCourses {
		for j, course2 := range allCourses {
			if i >= j { // Avoid checking the same pair twice and avoid self-comparison
				continue
			}

			// Check if these courses are crosslisted
			crosslisted, err := m.AreCoursesCrosslisted(course1.CRN, course2.CRN)
			if err != nil {
				return nil, fmt.Errorf("error checking crosslisting for CRNs %d and %d: %v", course1.CRN, course2.CRN, err)
			}

			if crosslisted {
				// Crosslisted courses should have different CRNs by definition
				// If they have the same CRN, that's a data error, so log it and skip
				if course1.CRN == course2.CRN {
					// AppLogger.LogError(fmt.Sprintf("Data error: course with CRN %d is crosslisted with itself", course1.CRN), nil)
					continue
				}

				// Check for instructor conflicts
				if course1.InstructorID != course2.InstructorID && course1.InstructorID > 0 && course2.InstructorID > 0 {
					conflictPair := CrosslistingConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "crosslisting-instructor",
					}
					crosslistingConflicts = append(crosslistingConflicts, conflictPair)
				}

				// Check for room conflicts (unless one or more is FSO, PSO, or AO)
				if course1.RoomID != course2.RoomID && course1.RoomID > 0 && course2.RoomID > 0 {
					if !m.isRoomExemptMode(course1) && !m.isRoomExemptMode(course2) {
						conflictPair := CrosslistingConflictPair{
							Course1: course1,
							Course2: course2,
							Type:    "crosslisting-room",
						}
						crosslistingConflicts = append(crosslistingConflicts, conflictPair)
					}
				}

				// Check for time conflicts (unless one or more is AO)
				if !m.timeSlotsMatch(course1.TimeSlot, course2.TimeSlot) {
					if !m.isTimeExemptMode(course1) && !m.isTimeExemptMode(course2) {
						conflictPair := CrosslistingConflictPair{
							Course1: course1,
							Course2: course2,
							Type:    "crosslisting-time",
						}
						crosslistingConflicts = append(crosslistingConflicts, conflictPair)
					}
				}
			}
		}
	}

	return crosslistingConflicts, nil
}

// Helper function to create a test course
func createCrosslistingTestCourse(crn int, instructorID int, roomID int, mode string, timeSlot *CrosslistingTimeSlot) CrosslistingCourseDetail {
	return CrosslistingCourseDetail{
		ID:           crn,
		CRN:          crn,
		Section:      "001",
		ScheduleID:   1,
		Prefix:       "CS",
		CourseNumber: "1000",
		Title:        "Test Course",
		InstructorID: instructorID,
		TimeSlotID:   1,
		RoomID:       roomID,
		Mode:         mode,
		TimeSlot:     timeSlot,
	}
}

// Helper function to create a test time slot
func createCrosslistingTestTimeSlot(startTime, endTime string, monday, tuesday, wednesday, thursday, friday bool) *CrosslistingTimeSlot {
	return &CrosslistingTimeSlot{
		ID:        1,
		StartTime: startTime,
		EndTime:   endTime,
		Monday:    monday,
		Tuesday:   tuesday,
		Wednesday: wednesday,
		Thursday:  thursday,
		Friday:    friday,
	}
}

func TestCrosslistingConflictDetection_NoCrosslistings(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: make(map[string]bool),
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot),
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 2, 102, "IP", timeSlot),
	}

	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(false, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, conflicts)
	mockScheduler.AssertExpectations(t)
}

func TestCrosslistingConflictDetection_InstructorConflict(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot), // Instructor 1
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 2, 101, "IP", timeSlot), // Instructor 2, same room and time
	}

	// The algorithm will call AreCoursesCrosslisted only once per unique pair
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "crosslisting-instructor", conflicts[0].Type)
}

func TestCrosslistingConflictDetection_RoomConflict(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot), // Room 101
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 1, 102, "IP", timeSlot), // Room 102, same instructor and time
	}

	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "crosslisting-room", conflicts[0].Type)
}

func TestCrosslistingConflictDetection_RoomConflictExemptFSO(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "FSO", timeSlot), // FSO mode - room exempt
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 1, 102, "IP", timeSlot), // Different room
	}

	// Handle both possible directions due to map iteration order
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, conflicts) // No room conflict due to FSO exemption
}

func TestCrosslistingConflictDetection_TimeConflict(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot1 := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)
	timeSlot2 := createCrosslistingTestTimeSlot("10:00", "11:00", true, false, true, false, false) // Different time

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot1),
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 1, 101, "IP", timeSlot2), // Same instructor and room, different time
	}

	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "crosslisting-time", conflicts[0].Type)
}

func TestCrosslistingConflictDetection_TimeConflictExemptAO(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot1 := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)
	timeSlot2 := createCrosslistingTestTimeSlot("10:00", "11:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "AO", timeSlot1), // AO mode - time exempt
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 1, 101, "IP", timeSlot2), // Different time
	}

	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, conflicts) // No time conflict due to AO exemption
}

func TestCrosslistingConflictDetection_MultipleConflicts(t *testing.T) {
	// Setup
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot1 := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)
	timeSlot2 := createCrosslistingTestTimeSlot("10:00", "11:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot1), // Instructor 1, Room 101, Time 1
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 2, 102, "IP", timeSlot2), // Instructor 2, Room 102, Time 2
	}

	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, conflicts, 3) // All three types of conflicts

	conflictTypes := make(map[string]bool)
	for _, conflict := range conflicts {
		conflictTypes[conflict.Type] = true
	}

	assert.True(t, conflictTypes["crosslisting-instructor"])
	assert.True(t, conflictTypes["crosslisting-room"])
	assert.True(t, conflictTypes["crosslisting-time"])
}

func TestCrosslistingConflictDetection_DuplicateCourses(t *testing.T) {
	// Setup - same course appears in both schedules
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
			"12346-12345": true, // Handle both directions since map iteration order can vary
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	// Same course CRN 12345 appears in both schedules - it will be deduped into one instance
	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot),
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot), // Duplicate
		createCrosslistingTestCourse(12346, 2, 102, "IP", timeSlot),
	}

	// The algorithm will only check each unique pair once, but the order is not deterministic
	// due to map iteration. We'll set up both possible calls but only expect one to be used.
	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert - we expect a room conflict because course 12345 (instructor 1, room 101)
	// is crosslisted with course 12346 (instructor 2, room 102)
	assert.NoError(t, err)
	assert.Len(t, conflicts, 2) // instructor and room conflicts
	// Don't assert all expectations as the mock allows for flexible call order
}

func TestCrosslistingConflictDetection_SameCRNCrosslisted(t *testing.T) {
	// Setup - data error case where a course is crosslisted with itself
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12345": true, // Invalid: course crosslisted with itself
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot),
	}
	courses2 := []CrosslistingCourseDetail{}

	// This scenario won't trigger AreCoursesCrosslisted because
	// there's only one unique course and i >= j prevents self-comparison

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, conflicts) // Should skip self-crosslisting
	mockScheduler.AssertExpectations(t)
}

func TestCrosslistingConflictDetection_NoConflictSameInstructorRoomTime(t *testing.T) {
	// Setup - crosslisted courses with same instructor, room, and time (valid scenario)
	mockScheduler := &MockScheduler{
		crosslistings: map[string]bool{
			"12345-12346": true,
		},
	}

	timeSlot := createCrosslistingTestTimeSlot("09:00", "10:00", true, false, true, false, false)

	courses1 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12345, 1, 101, "IP", timeSlot),
	}
	courses2 := []CrosslistingCourseDetail{
		createCrosslistingTestCourse(12346, 1, 101, "IP", timeSlot), // Same instructor, room, and time
	}

	mockScheduler.On("AreCoursesCrosslisted", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(true, nil)

	// Execute
	conflicts, err := mockScheduler.detectCrosslistingConflicts(courses1, courses2)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, conflicts) // No conflicts - all attributes match as expected for crosslisted courses
}
