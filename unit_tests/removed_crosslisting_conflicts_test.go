package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test data structures for removed courses and crosslisting conflicts
type RemovedCourseDetail struct {
	ID                  int
	CRN                 int
	Section             string
	ScheduleID          int
	Prefix              string
	CourseNumber        string
	Title               string
	InstructorID        int
	InstructorFirstName string
	InstructorLastName  string
	TimeSlotID          int
	RoomID              int
	Mode                string
	Status              string
	Lab                 bool
	TimeSlot            *RemovedTimeSlot
}

type RemovedTimeSlot struct {
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

type RemovedConflictPair struct {
	Course1 RemovedCourseDetail
	Course2 RemovedCourseDetail
	Type    string
}

type RemovedConflictReport struct {
	InstructorConflicts   []RemovedConflictPair
	RoomConflicts         []RemovedConflictPair
	CrosslistingConflicts []RemovedConflictPair
	CourseConflicts       []RemovedConflictPair
}

// MockRemovedScheduler implements the scheduler interface for testing removed course logic
type MockRemovedScheduler struct {
	mock.Mock
	crosslistings map[string]bool
	prerequisites map[string][]string
}

func (m *MockRemovedScheduler) AreCoursesCrosslisted(crn1, crn2 int) (bool, error) {
	args := m.Called(crn1, crn2)
	return args.Bool(0), args.Error(1)
}

func (m *MockRemovedScheduler) areCoursesOnSamePrerequisiteChain(prefix1, courseNum1, prefix2, courseNum2 string) (bool, error) {
	args := m.Called(prefix1, courseNum1, prefix2, courseNum2)
	return args.Bool(0), args.Error(1)
}

// Helper functions to mimic the actual conflict detection logic

// timeSlotsOverlap checks if two time slots overlap in both time and days
func (m *MockRemovedScheduler) timeSlotsOverlap(slot1, slot2 *RemovedTimeSlot) bool {
	if slot1 == nil || slot2 == nil {
		return false
	}

	// Check if days overlap
	daysOverlap := (slot1.Monday && slot2.Monday) ||
		(slot1.Tuesday && slot2.Tuesday) ||
		(slot1.Wednesday && slot2.Wednesday) ||
		(slot1.Thursday && slot2.Thursday) ||
		(slot1.Friday && slot2.Friday)

	if !daysOverlap {
		return false
	}

	// Check if times overlap
	return m.timeRangesOverlap(slot1.StartTime, slot1.EndTime, slot2.StartTime, slot2.EndTime)
}

func (m *MockRemovedScheduler) timeRangesOverlap(start1, end1, start2, end2 string) bool {
	// Simple string comparison for testing purposes
	// In real implementation, this would parse times properly
	return !(end1 <= start2 || end2 <= start1)
}

func (m *MockRemovedScheduler) timeSlotsMatch(slot1, slot2 *RemovedTimeSlot) bool {
	if slot1 == nil || slot2 == nil {
		return slot1 == slot2
	}

	return slot1.StartTime == slot2.StartTime &&
		slot1.EndTime == slot2.EndTime &&
		slot1.Monday == slot2.Monday &&
		slot1.Tuesday == slot2.Tuesday &&
		slot1.Wednesday == slot2.Wednesday &&
		slot1.Thursday == slot2.Thursday &&
		slot1.Friday == slot2.Friday
}

func (m *MockRemovedScheduler) isRoomExemptMode(course RemovedCourseDetail) bool {
	return course.Mode == "FSO" || course.Mode == "PSO" || course.Mode == "AO"
}

func (m *MockRemovedScheduler) isTimeExemptMode(course RemovedCourseDetail) bool {
	return course.Mode == "AO"
}

func (m *MockRemovedScheduler) isFSOPSOException(course1, course2 RemovedCourseDetail) bool {
	return (course1.Mode == "FSO" || course1.Mode == "PSO") &&
		(course2.Mode == "FSO" || course2.Mode == "PSO")
}

// Main conflict detection function that mimics the actual DetectConflictsBetweenSchedules
func (m *MockRemovedScheduler) DetectConflictsBetweenSchedules(courses1, courses2 []RemovedCourseDetail) (*RemovedConflictReport, error) {
	report := &RemovedConflictReport{
		InstructorConflicts:   []RemovedConflictPair{},
		RoomConflicts:         []RemovedConflictPair{},
		CrosslistingConflicts: []RemovedConflictPair{},
		CourseConflicts:       []RemovedConflictPair{},
	}

	// Test instructor conflicts (should skip removed courses)
	for _, course1 := range courses1 {
		for _, course2 := range courses2 {
			// Skip courses with "Removed" status - they cannot conflict with any other course
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Check instructor conflicts
			if course1.InstructorID > 0 && course2.InstructorID > 0 && course1.InstructorID == course2.InstructorID {
				if !m.isFSOPSOException(course1, course2) && m.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
					report.InstructorConflicts = append(report.InstructorConflicts, RemovedConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Instructor",
					})
				}
			}

			// Check room conflicts
			if course1.RoomID > 0 && course2.RoomID > 0 && course1.RoomID == course2.RoomID {
				if !m.isRoomExemptMode(course1) && !m.isRoomExemptMode(course2) && m.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
					report.RoomConflicts = append(report.RoomConflicts, RemovedConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Room",
					})
				}
			}
		}
	}

	// Test crosslisting conflicts
	crosslistingConflicts, err := m.detectCrosslistingConflicts(courses1, courses2)
	if err != nil {
		return nil, err
	}
	report.CrosslistingConflicts = crosslistingConflicts

	// Test course conflicts
	courseConflicts, err := m.detectCourseConflicts(courses1, courses2)
	if err != nil {
		return nil, err
	}
	report.CourseConflicts = courseConflicts

	return report, nil
}

// detectCrosslistingConflicts - should skip removed courses
func (m *MockRemovedScheduler) detectCrosslistingConflicts(courses1, courses2 []RemovedCourseDetail) ([]RemovedConflictPair, error) {
	var crosslistingConflicts []RemovedConflictPair

	// Create map of all courses by CRN
	courseMap := make(map[int]RemovedCourseDetail)
	for _, course := range courses1 {
		courseMap[course.CRN] = course
	}
	for _, course := range courses2 {
		if _, exists := courseMap[course.CRN]; !exists {
			courseMap[course.CRN] = course
		}
	}

	var allCourses []RemovedCourseDetail
	for _, course := range courseMap {
		allCourses = append(allCourses, course)
	}

	for i := 0; i < len(allCourses); i++ {
		for j := i + 1; j < len(allCourses); j++ {
			course1 := allCourses[i]
			course2 := allCourses[j]

			// Skip courses with "Removed" status - they cannot conflict with any other course
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Check if these courses are crosslisted
			crosslisted, err := m.AreCoursesCrosslisted(course1.CRN, course2.CRN)
			if err != nil {
				return nil, err
			}

			if crosslisted {
				// Crosslisted courses should not have time conflicts (unless AO mode)
				if course1.TimeSlot != nil && course2.TimeSlot != nil {
					if !m.isTimeExemptMode(course1) && !m.isTimeExemptMode(course2) && !m.timeSlotsMatch(course1.TimeSlot, course2.TimeSlot) {
						crosslistingConflicts = append(crosslistingConflicts, RemovedConflictPair{
							Course1: course1,
							Course2: course2,
							Type:    "Crosslisting",
						})
					}
				}
			}
		}
	}

	return crosslistingConflicts, nil
}

// detectCourseConflicts - should skip removed courses
func (m *MockRemovedScheduler) detectCourseConflicts(courses1, courses2 []RemovedCourseDetail) ([]RemovedConflictPair, error) {
	var courseConflicts []RemovedConflictPair

	for _, course1 := range courses1 {
		for _, course2 := range courses2 {
			// Skip courses with "Removed" status - they cannot conflict with any other course
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Only check courses with the same prefix
			if course1.Prefix != course2.Prefix {
				continue
			}

			// Check if courses are in the same range (1000-1999, 2000-2999, etc.)
			if m.isInSameCourseRange(course1.CourseNumber, course2.CourseNumber) {
				// Check for exceptions (crosslisted or prerequisite chain)
				isException, err := m.isCourseConflictException(course1, course2)
				if err != nil {
					return nil, err
				}

				if !isException {
					courseConflicts = append(courseConflicts, RemovedConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Course",
					})
				}
			}
		}
	}

	return courseConflicts, nil
}

func (m *MockRemovedScheduler) isInSameCourseRange(courseNum1, courseNum2 string) bool {
	// Simple implementation for testing
	if len(courseNum1) >= 4 && len(courseNum2) >= 4 {
		return courseNum1[0] == courseNum2[0] // Same thousand range
	}
	return false
}

func (m *MockRemovedScheduler) isCourseConflictException(course1, course2 RemovedCourseDetail) (bool, error) {
	// Check if courses are crosslisted
	crosslisted, err := m.AreCoursesCrosslisted(course1.CRN, course2.CRN)
	if err != nil {
		return false, err
	}
	if crosslisted {
		return true, nil
	}

	// Check if courses are on same prerequisite chain
	onSameChain, err := m.areCoursesOnSamePrerequisiteChain(course1.Prefix, course1.CourseNumber, course2.Prefix, course2.CourseNumber)
	if err != nil {
		return false, err
	}

	return onSameChain, nil
}

// Test cases for "Removed" status courses

func TestRemovedCourses_InstructorConflicts_ShouldBeSkipped(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Create test time slot that overlaps
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create courses with same instructor and overlapping time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              1,
			Mode:                "Traditional",
			Status:              "Removed", // This course has "Removed" status
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "CS",
			CourseNumber:        "2160",
			Title:               "Computer Science II",
			InstructorID:        100, // Same instructor
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          2,
			RoomID:              2,
			Mode:                "Traditional",
			Status:              "Scheduled", // This course is normal
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.InstructorConflicts, "Removed courses should not create instructor conflicts")
	assert.Empty(t, report.RoomConflicts, "Removed courses should not create room conflicts")
}

func TestRemovedCourses_RoomConflicts_ShouldBeSkipped(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Create test time slot that overlaps
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create courses with same room and overlapping time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200, // Same room
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "MATH",
			CourseNumber:        "1220",
			Title:               "Calculus I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          2,
			RoomID:              200, // Same room
			Mode:                "Traditional",
			Status:              "Removed", // This course has "Removed" status
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.RoomConflicts, "Removed courses should not create room conflicts")
	assert.Empty(t, report.InstructorConflicts, "Removed courses should not create instructor conflicts")
}

func TestRemovedCourses_CrosslistingConflicts_ShouldBeSkipped(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Set up mock to return that courses are crosslisted
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)

	// Create different time slots for crosslisted courses (should normally conflict)
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create crosslisted courses with different time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "Traditional",
			Status:              "Removed", // This course has "Removed" status
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          2,
			RoomID:              201,
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.CrosslistingConflicts, "Removed courses should not create crosslisting conflicts")
}

func TestRemovedCourses_CourseConflicts_ShouldBeSkipped(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Set up mocks - courses are not crosslisted and not on prerequisite chain
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(false, nil)
	mockScheduler.On("areCoursesOnSamePrerequisiteChain", "CS", "2150", "CS", "2160").Return(false, nil)

	// Create courses in same range that would normally conflict
	courses1 := []RemovedCourseDetail{
		{
			ID:           1,
			CRN:          12345,
			Section:      "001",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150", // Same 2000 range
			Title:        "Computer Science I",
			InstructorID: 100,
			TimeSlotID:   1,
			RoomID:       200,
			Mode:         "Traditional",
			Status:       "Removed", // This course has "Removed" status
			Lab:          false,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:           2,
			CRN:          12346,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2160", // Same 2000 range, would normally conflict
			Title:        "Computer Science II",
			InstructorID: 101,
			TimeSlotID:   2,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.CourseConflicts, "Removed courses should not create course conflicts")
}

// Test cases for cross-listed courses behavior

func TestCrosslistedCourses_SameTimeSlot_ShouldNotConflict(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Set up mock to return that courses are crosslisted
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)

	// Create identical time slots for crosslisted courses
	timeSlot := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create crosslisted courses with identical time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "Traditional",
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          1,
			RoomID:              201,
			Mode:                "Traditional",
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot, // Same time slot
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.CrosslistingConflicts, "Crosslisted courses with same time slots should not conflict")
}

func TestCrosslistedCourses_DifferentTimeSlots_ShouldConflict(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Set up mock to return that courses are crosslisted
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)

	// Create different time slots for crosslisted courses
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create crosslisted courses with different time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "Traditional",
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          2,
			RoomID:              201,
			Mode:                "Traditional",
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot2, // Different time slot
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.CrosslistingConflicts, 1, "Crosslisted courses with different time slots should conflict")
	assert.Equal(t, "Crosslisting", report.CrosslistingConflicts[0].Type)
	assert.Equal(t, 12345, report.CrosslistingConflicts[0].Course1.CRN)
	assert.Equal(t, 12346, report.CrosslistingConflicts[0].Course2.CRN)
}

func TestCrosslistedCourses_AOMode_ShouldNotConflict(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Set up mock to return that courses are crosslisted
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)

	// Create different time slots for crosslisted courses
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create crosslisted courses with different time slots but one in AO mode
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "AO", // Arrangement Only mode - exempt from time conflicts
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          2,
			RoomID:              201,
			Mode:                "Traditional",
			Status:              "Scheduled",
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.CrosslistingConflicts, "Crosslisted courses in AO mode should not create conflicts")
}

// Test cases for normal (non-removed, non-crosslisted) courses that should still conflict

func TestNormalCourses_InstructorConflicts_ShouldStillWork(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Create overlapping time slots
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create normal courses with same instructor and overlapping time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "MATH",
			CourseNumber:        "1220",
			Title:               "Calculus I",
			InstructorID:        100, // Same instructor
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          2,
			RoomID:              201,
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.InstructorConflicts, 1, "Normal courses with same instructor and overlapping times should conflict")
	assert.Equal(t, "Instructor", report.InstructorConflicts[0].Type)
	assert.Equal(t, 100, report.InstructorConflicts[0].Course1.InstructorID)
	assert.Equal(t, 100, report.InstructorConflicts[0].Course2.InstructorID)
}

func TestNormalCourses_RoomConflicts_ShouldStillWork(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Create overlapping time slots
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create normal courses with same room and overlapping time slots
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200, // Same room
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "MATH",
			CourseNumber:        "1220",
			Title:               "Calculus I",
			InstructorID:        101,
			InstructorFirstName: "Jane",
			InstructorLastName:  "Smith",
			TimeSlotID:          2,
			RoomID:              200, // Same room
			Mode:                "Traditional",
			Status:              "Scheduled", // Normal course
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.RoomConflicts, 1, "Normal courses with same room and overlapping times should conflict")
	assert.Equal(t, "Room", report.RoomConflicts[0].Type)
	assert.Equal(t, 200, report.RoomConflicts[0].Course1.RoomID)
	assert.Equal(t, 200, report.RoomConflicts[0].Course2.RoomID)
}

// Test edge cases

func TestBothCoursesRemoved_ShouldNotConflict(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Create overlapping time slots
	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create courses where both are removed
	courses1 := []RemovedCourseDetail{
		{
			ID:                  1,
			CRN:                 12345,
			Section:             "001",
			ScheduleID:          1,
			Prefix:              "CS",
			CourseNumber:        "2150",
			Title:               "Computer Science I",
			InstructorID:        100,
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          1,
			RoomID:              200,
			Mode:                "Traditional",
			Status:              "Removed", // Both courses are removed
			Lab:                 false,
			TimeSlot:            timeSlot1,
		},
	}

	courses2 := []RemovedCourseDetail{
		{
			ID:                  2,
			CRN:                 12346,
			Section:             "002",
			ScheduleID:          2,
			Prefix:              "MATH",
			CourseNumber:        "1220",
			Title:               "Calculus I",
			InstructorID:        100, // Same instructor
			InstructorFirstName: "John",
			InstructorLastName:  "Doe",
			TimeSlotID:          2,
			RoomID:              200, // Same room
			Mode:                "Traditional",
			Status:              "Removed", // Both courses are removed
			Lab:                 false,
			TimeSlot:            timeSlot2,
		},
	}

	// Execute the conflict detection
	report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Empty(t, report.InstructorConflicts, "Both removed courses should not create any conflicts")
	assert.Empty(t, report.RoomConflicts, "Both removed courses should not create any conflicts")
	assert.Empty(t, report.CrosslistingConflicts, "Both removed courses should not create any conflicts")
	assert.Empty(t, report.CourseConflicts, "Both removed courses should not create any conflicts")
}

func TestMixedStatusCombinations(t *testing.T) {
	mockScheduler := new(MockRemovedScheduler)

	// Test various status combinations to ensure only "Removed" is excluded
	testCases := []struct {
		name           string
		status1        string
		status2        string
		shouldConflict bool
	}{
		{"Scheduled vs Scheduled", "Scheduled", "Scheduled", true},
		{"Scheduled vs Removed", "Scheduled", "Removed", false},
		{"Removed vs Scheduled", "Removed", "Scheduled", false},
		{"Removed vs Removed", "Removed", "Removed", false},
		{"Scheduled vs Cancelled", "Scheduled", "Cancelled", true},
		{"Cancelled vs Removed", "Cancelled", "Removed", false},
		{"Active vs Removed", "Active", "Removed", false},
	}

	timeSlot1 := &RemovedTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &RemovedTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			courses1 := []RemovedCourseDetail{
				{
					ID:                  1,
					CRN:                 12345,
					Section:             "001",
					ScheduleID:          1,
					Prefix:              "CS",
					CourseNumber:        "2150",
					Title:               "Computer Science I",
					InstructorID:        100,
					InstructorFirstName: "John",
					InstructorLastName:  "Doe",
					TimeSlotID:          1,
					RoomID:              200,
					Mode:                "Traditional",
					Status:              tc.status1,
					Lab:                 false,
					TimeSlot:            timeSlot1,
				},
			}

			courses2 := []RemovedCourseDetail{
				{
					ID:                  2,
					CRN:                 12346,
					Section:             "002",
					ScheduleID:          2,
					Prefix:              "MATH",
					CourseNumber:        "1220",
					Title:               "Calculus I",
					InstructorID:        100, // Same instructor
					InstructorFirstName: "John",
					InstructorLastName:  "Doe",
					TimeSlotID:          2,
					RoomID:              200, // Same room
					Mode:                "Traditional",
					Status:              tc.status2,
					Lab:                 false,
					TimeSlot:            timeSlot2,
				},
			}

			// Execute the conflict detection
			report, err := mockScheduler.DetectConflictsBetweenSchedules(courses1, courses2)

			// Assertions
			assert.NoError(t, err)
			assert.NotNil(t, report)

			if tc.shouldConflict {
				assert.NotEmpty(t, report.InstructorConflicts, "Expected instructor conflicts for %s", tc.name)
				assert.NotEmpty(t, report.RoomConflicts, "Expected room conflicts for %s", tc.name)
			} else {
				assert.Empty(t, report.InstructorConflicts, "Expected no instructor conflicts for %s", tc.name)
				assert.Empty(t, report.RoomConflicts, "Expected no room conflicts for %s", tc.name)
			}
		})
	}
}
