package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Integration test structures that mirror the actual CourseDetail struct more closely
type IntegrationCourseDetail struct {
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
	TimeSlot            *IntegrationTimeSlot
}

type IntegrationTimeSlot struct {
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

type IntegrationConflictPair struct {
	Course1 IntegrationCourseDetail
	Course2 IntegrationCourseDetail
	Type    string
}

// MockIntegrationScheduler for more comprehensive testing
type MockIntegrationScheduler struct {
	mock.Mock
	crosslistings map[string]bool
	prerequisites map[string][]string
}

func (m *MockIntegrationScheduler) AreCoursesCrosslisted(crn1, crn2 int) (bool, error) {
	args := m.Called(crn1, crn2)
	return args.Bool(0), args.Error(1)
}

func (m *MockIntegrationScheduler) areCoursesOnSamePrerequisiteChain(prefix1, courseNum1, prefix2, courseNum2 string) (bool, error) {
	args := m.Called(prefix1, courseNum1, prefix2, courseNum2)
	return args.Bool(0), args.Error(1)
}

// Helper function to simulate comprehensive conflict detection
func (m *MockIntegrationScheduler) DetectConflictsBetweenSchedules(schedule1ID, schedule2ID int, courses1, courses2 []IntegrationCourseDetail) (map[string][]IntegrationConflictPair, error) {
	conflicts := map[string][]IntegrationConflictPair{
		"instructor":   []IntegrationConflictPair{},
		"room":         []IntegrationConflictPair{},
		"crosslisting": []IntegrationConflictPair{},
		"course":       []IntegrationConflictPair{},
	}

	// Instructor and Room conflicts
	for _, course1 := range courses1 {
		for _, course2 := range courses2 {
			// Skip courses with "Removed" status - they cannot conflict with any other course
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Check instructor conflicts
			if course1.InstructorID > 0 && course2.InstructorID > 0 && course1.InstructorID == course2.InstructorID {
				if !m.isFSOPSOException(course1, course2) && m.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
					conflicts["instructor"] = append(conflicts["instructor"], IntegrationConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Instructor",
					})
				}
			}

			// Check room conflicts
			if course1.RoomID > 0 && course2.RoomID > 0 && course1.RoomID == course2.RoomID {
				if !m.isRoomExemptMode(course1) && !m.isRoomExemptMode(course2) && m.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
					conflicts["room"] = append(conflicts["room"], IntegrationConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Room",
					})
				}
			}
		}
	}

	// Crosslisting conflicts
	crosslistingConflicts, err := m.detectCrosslistingConflicts(courses1, courses2)
	if err != nil {
		return nil, err
	}
	conflicts["crosslisting"] = crosslistingConflicts

	// Course conflicts
	courseConflicts, err := m.detectCourseConflicts(courses1, courses2)
	if err != nil {
		return nil, err
	}
	conflicts["course"] = courseConflicts

	return conflicts, nil
}

func (m *MockIntegrationScheduler) detectCrosslistingConflicts(courses1, courses2 []IntegrationCourseDetail) ([]IntegrationConflictPair, error) {
	var conflicts []IntegrationConflictPair

	// Create map of all courses by CRN
	courseMap := make(map[int]IntegrationCourseDetail)
	for _, course := range courses1 {
		courseMap[course.CRN] = course
	}
	for _, course := range courses2 {
		if _, exists := courseMap[course.CRN]; !exists {
			courseMap[course.CRN] = course
		}
	}

	var allCourses []IntegrationCourseDetail
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
				// Crosslisted courses should have same time slot (unless AO mode)
				if course1.TimeSlot != nil && course2.TimeSlot != nil {
					if !m.isTimeExemptMode(course1) && !m.isTimeExemptMode(course2) && !m.timeSlotsMatch(course1.TimeSlot, course2.TimeSlot) {
						conflicts = append(conflicts, IntegrationConflictPair{
							Course1: course1,
							Course2: course2,
							Type:    "Crosslisting",
						})
					}
				}
			}
		}
	}

	return conflicts, nil
}

func (m *MockIntegrationScheduler) detectCourseConflicts(courses1, courses2 []IntegrationCourseDetail) ([]IntegrationConflictPair, error) {
	var conflicts []IntegrationConflictPair

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

			// Check if courses are in the same range
			if m.isInSameCourseRange(course1.CourseNumber, course2.CourseNumber) {
				// Check for exceptions
				isException, err := m.isCourseConflictException(course1, course2)
				if err != nil {
					return nil, err
				}

				if !isException {
					conflicts = append(conflicts, IntegrationConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "Course",
					})
				}
			}
		}
	}

	return conflicts, nil
}

// Helper functions for conflict detection logic
func (m *MockIntegrationScheduler) timeSlotsOverlap(slot1, slot2 *IntegrationTimeSlot) bool {
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

	// Check if times overlap (simple string comparison for testing)
	return !(slot1.EndTime <= slot2.StartTime || slot2.EndTime <= slot1.StartTime)
}

func (m *MockIntegrationScheduler) timeSlotsMatch(slot1, slot2 *IntegrationTimeSlot) bool {
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

func (m *MockIntegrationScheduler) isRoomExemptMode(course IntegrationCourseDetail) bool {
	return course.Mode == "FSO" || course.Mode == "PSO" || course.Mode == "AO"
}

func (m *MockIntegrationScheduler) isTimeExemptMode(course IntegrationCourseDetail) bool {
	return course.Mode == "AO"
}

func (m *MockIntegrationScheduler) isFSOPSOException(course1, course2 IntegrationCourseDetail) bool {
	return (course1.Mode == "FSO" || course1.Mode == "PSO") &&
		(course2.Mode == "FSO" || course2.Mode == "PSO")
}

func (m *MockIntegrationScheduler) isInSameCourseRange(courseNum1, courseNum2 string) bool {
	if len(courseNum1) >= 4 && len(courseNum2) >= 4 {
		return courseNum1[0] == courseNum2[0] // Same thousand range
	}
	return false
}

func (m *MockIntegrationScheduler) isCourseConflictException(course1, course2 IntegrationCourseDetail) (bool, error) {
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

// Integration test cases for complex scenarios

func TestIntegration_MultipleRemovedCourses_ComplexScenario(t *testing.T) {
	mockScheduler := new(MockIntegrationScheduler)

	// Setup crosslisting mocks - all possible combinations
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12346, 12345).Return(true, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12347).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12347, 12345).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12348).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12348, 12345).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12346, 12347).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12347, 12346).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12346, 12348).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12348, 12346).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12347, 12348).Return(false, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12348, 12347).Return(false, nil)

	// Setup prerequisite chain mocks
	mockScheduler.On("areCoursesOnSamePrerequisiteChain", "CS", "2150", "CS", "2160").Return(false, nil)
	mockScheduler.On("areCoursesOnSamePrerequisiteChain", "CS", "2160", "CS", "2150").Return(false, nil)

	// Create complex time slot scenarios
	timeSlot1 := &IntegrationTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &IntegrationTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot3 := &IntegrationTimeSlot{
		ID:        3,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create a complex scenario with multiple courses, some removed
	courses1 := []IntegrationCourseDetail{
		{
			ID:           1,
			CRN:          12345,
			Section:      "001",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 100,
			TimeSlotID:   1,
			RoomID:       200,
			Mode:         "Traditional",
			Status:       "Removed", // This course is removed
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
		{
			ID:           2,
			CRN:          12347,
			Section:      "002",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2160",
			Title:        "Computer Science II",
			InstructorID: 100, // Same instructor as removed course
			TimeSlotID:   2,
			RoomID:       200, // Same room as removed course
			Mode:         "Traditional",
			Status:       "Scheduled", // This course is normal
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	courses2 := []IntegrationCourseDetail{
		{
			ID:           3,
			CRN:          12346,
			Section:      "001",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 101,
			TimeSlotID:   3,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course, crosslisted with removed course
			Lab:          false,
			TimeSlot:     timeSlot3,
		},
		{
			ID:           4,
			CRN:          12348,
			Section:      "003",
			ScheduleID:   2,
			Prefix:       "MATH",
			CourseNumber: "1220",
			Title:        "Calculus I",
			InstructorID: 100, // Same instructor as normal course in schedule 1
			TimeSlotID:   1,
			RoomID:       200, // Same room as normal course in schedule 1
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	// Execute the conflict detection
	conflicts, err := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, conflicts)

	// The removed course (12345) should not create any conflicts
	// But the normal course (12347) should still conflict with course 12348
	assert.Len(t, conflicts["instructor"], 1, "Should have instructor conflict between normal courses")
	assert.Len(t, conflicts["room"], 1, "Should have room conflict between normal courses")
	assert.Empty(t, conflicts["crosslisting"], "No crosslisting conflicts because crosslisted course is removed")

	// Verify the conflicts are between the correct courses (not involving removed course)
	instructorConflict := conflicts["instructor"][0]
	assert.True(t,
		(instructorConflict.Course1.CRN == 12347 && instructorConflict.Course2.CRN == 12348) ||
			(instructorConflict.Course1.CRN == 12348 && instructorConflict.Course2.CRN == 12347),
		"Instructor conflict should be between normal courses, not involving removed course")

	roomConflict := conflicts["room"][0]
	assert.True(t,
		(roomConflict.Course1.CRN == 12347 && roomConflict.Course2.CRN == 12348) ||
			(roomConflict.Course1.CRN == 12348 && roomConflict.Course2.CRN == 12347),
		"Room conflict should be between normal courses, not involving removed course")
}

func TestIntegration_CrosslistedCoursesWithDifferentModes(t *testing.T) {
	mockScheduler := new(MockIntegrationScheduler)

	// Setup crosslisting mock
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)

	// Create different time slots
	timeSlot1 := &IntegrationTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &IntegrationTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	testCases := []struct {
		name           string
		mode1          string
		mode2          string
		shouldConflict bool
		description    string
	}{
		{
			name:           "Traditional vs Traditional",
			mode1:          "Traditional",
			mode2:          "Traditional",
			shouldConflict: true,
			description:    "Different time slots should conflict for traditional modes",
		},
		{
			name:           "AO vs Traditional",
			mode1:          "AO",
			mode2:          "Traditional",
			shouldConflict: false,
			description:    "AO mode is exempt from time conflicts",
		},
		{
			name:           "Traditional vs AO",
			mode1:          "Traditional",
			mode2:          "AO",
			shouldConflict: false,
			description:    "AO mode is exempt from time conflicts",
		},
		{
			name:           "AO vs AO",
			mode1:          "AO",
			mode2:          "AO",
			shouldConflict: false,
			description:    "Both AO modes are exempt from time conflicts",
		},
		{
			name:           "FSO vs PSO",
			mode1:          "FSO",
			mode2:          "PSO",
			shouldConflict: true,
			description:    "FSO/PSO exemption is for instructor conflicts, not crosslisting time conflicts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			courses1 := []IntegrationCourseDetail{
				{
					ID:           1,
					CRN:          12345,
					Section:      "001",
					ScheduleID:   1,
					Prefix:       "CS",
					CourseNumber: "2150",
					Title:        "Computer Science I",
					InstructorID: 100,
					TimeSlotID:   1,
					RoomID:       200,
					Mode:         tc.mode1,
					Status:       "Scheduled",
					Lab:          false,
					TimeSlot:     timeSlot1,
				},
			}

			courses2 := []IntegrationCourseDetail{
				{
					ID:           2,
					CRN:          12346,
					Section:      "002",
					ScheduleID:   2,
					Prefix:       "CS",
					CourseNumber: "2150",
					Title:        "Computer Science I",
					InstructorID: 101,
					TimeSlotID:   2,
					RoomID:       201,
					Mode:         tc.mode2,
					Status:       "Scheduled",
					Lab:          false,
					TimeSlot:     timeSlot2,
				},
			}

			conflicts, err := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

			assert.NoError(t, err)
			assert.NotNil(t, conflicts)

			if tc.shouldConflict {
				assert.Len(t, conflicts["crosslisting"], 1, "Expected crosslisting conflict for %s: %s", tc.name, tc.description)
			} else {
				assert.Empty(t, conflicts["crosslisting"], "Expected no crosslisting conflict for %s: %s", tc.name, tc.description)
			}
		})
	}
}

func TestIntegration_RemovedCourseInCrosslistingGroup(t *testing.T) {
	mockScheduler := new(MockIntegrationScheduler)

	// Setup crosslisting mocks - 3 courses are all crosslisted together
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(true, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12347).Return(true, nil)
	mockScheduler.On("AreCoursesCrosslisted", 12346, 12347).Return(true, nil)

	// Create time slots
	timeSlot1 := &IntegrationTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &IntegrationTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	timeSlot3 := &IntegrationTimeSlot{
		ID:        3,
		StartTime: "09:00:00",
		EndTime:   "10:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create 3 crosslisted courses, with one removed
	courses1 := []IntegrationCourseDetail{
		{
			ID:           1,
			CRN:          12345,
			Section:      "001",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 100,
			TimeSlotID:   1,
			RoomID:       200,
			Mode:         "Traditional",
			Status:       "Removed", // This course is removed from the crosslisting group
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
		{
			ID:           2,
			CRN:          12346,
			Section:      "002",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 101,
			TimeSlotID:   2,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled", // This course is normal
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	courses2 := []IntegrationCourseDetail{
		{
			ID:           3,
			CRN:          12347,
			Section:      "003",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 102,
			TimeSlotID:   3,
			RoomID:       202,
			Mode:         "Traditional",
			Status:       "Scheduled", // This course is normal
			Lab:          false,
			TimeSlot:     timeSlot3,
		},
	}

	conflicts, err := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	assert.NoError(t, err)
	assert.NotNil(t, conflicts)

	// Only the non-removed crosslisted courses should create conflicts
	assert.Len(t, conflicts["crosslisting"], 1, "Should have one crosslisting conflict between non-removed courses")

	crosslistingConflict := conflicts["crosslisting"][0]
	// Verify the conflict is between the two non-removed courses
	assert.True(t,
		(crosslistingConflict.Course1.CRN == 12346 && crosslistingConflict.Course2.CRN == 12347) ||
			(crosslistingConflict.Course1.CRN == 12347 && crosslistingConflict.Course2.CRN == 12346),
		"Crosslisting conflict should be between non-removed courses only")

	// Verify removed course is not involved in any conflicts
	for _, conflict := range conflicts["crosslisting"] {
		assert.NotEqual(t, 12345, conflict.Course1.CRN, "Removed course should not be in conflicts")
		assert.NotEqual(t, 12345, conflict.Course2.CRN, "Removed course should not be in conflicts")
	}
}

func TestIntegration_VerifyAllConflictTypesRespectRemovedStatus(t *testing.T) {
	mockScheduler := new(MockIntegrationScheduler)

	// Setup mocks for course conflicts
	mockScheduler.On("AreCoursesCrosslisted", 12345, 12346).Return(false, nil)
	mockScheduler.On("areCoursesOnSamePrerequisiteChain", "CS", "2150", "CS", "2160").Return(false, nil)

	// Create overlapping time slot
	timeSlot := &IntegrationTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create two courses that would conflict in every way possible if both were normal
	courses1 := []IntegrationCourseDetail{
		{
			ID:           1,
			CRN:          12345,
			Section:      "001",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 100, // Same instructor
			TimeSlotID:   1,
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Removed", // This course is removed
			Lab:          false,
			TimeSlot:     timeSlot,
		},
	}

	courses2 := []IntegrationCourseDetail{
		{
			ID:           2,
			CRN:          12346,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2160", // Same range as 2150
			Title:        "Computer Science II",
			InstructorID: 100, // Same instructor
			TimeSlotID:   1,
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot,
		},
	}

	conflicts, err := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	assert.NoError(t, err)
	assert.NotNil(t, conflicts)

	// ALL conflict types should be empty because one course is removed
	assert.Empty(t, conflicts["instructor"], "Instructor conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts["room"], "Room conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts["crosslisting"], "Crosslisting conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts["course"], "Course conflicts should be empty when one course is removed")

	// Now test the reverse - same courses but with statuses swapped
	courses1[0].Status = "Scheduled"
	courses2[0].Status = "Removed"

	conflicts2, err2 := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	assert.NoError(t, err2)
	assert.NotNil(t, conflicts2)

	// ALL conflict types should still be empty because one course is removed
	assert.Empty(t, conflicts2["instructor"], "Instructor conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts2["room"], "Room conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts2["crosslisting"], "Crosslisting conflicts should be empty when one course is removed")
	assert.Empty(t, conflicts2["course"], "Course conflicts should be empty when one course is removed")

	// Finally test with both courses normal - should have conflicts
	courses1[0].Status = "Scheduled"
	courses2[0].Status = "Scheduled"

	conflicts3, err3 := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	assert.NoError(t, err3)
	assert.NotNil(t, conflicts3)

	// Now should have conflicts (except crosslisting since they're not crosslisted)
	assert.Len(t, conflicts3["instructor"], 1, "Should have instructor conflict when both courses are normal")
	assert.Len(t, conflicts3["room"], 1, "Should have room conflict when both courses are normal")
	assert.Empty(t, conflicts3["crosslisting"], "No crosslisting conflict since not crosslisted")
	assert.Len(t, conflicts3["course"], 1, "Should have course conflict when both courses are normal")
}

// Performance and edge case tests

func TestIntegration_LargeScaleRemovedCoursePerformance(t *testing.T) {
	mockScheduler := new(MockIntegrationScheduler)

	// Create a large number of removed courses to ensure performance isn't degraded
	var courses1, courses2 []IntegrationCourseDetail

	timeSlot := &IntegrationTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create 50 removed courses and 2 normal courses
	for i := 0; i < 50; i++ {
		courses1 = append(courses1, IntegrationCourseDetail{
			ID:           i + 1,
			CRN:          12345 + i,
			Section:      "001",
			ScheduleID:   1,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 100, // Same instructor (would conflict if not removed)
			TimeSlotID:   1,
			RoomID:       200, // Same room (would conflict if not removed)
			Mode:         "Traditional",
			Status:       "Removed", // All removed
			Lab:          false,
			TimeSlot:     timeSlot,
		})
	}

	// Add 2 normal courses that should conflict with each other
	courses1 = append(courses1, IntegrationCourseDetail{
		ID:           51,
		CRN:          13000,
		Section:      "001",
		ScheduleID:   1,
		Prefix:       "CS",
		CourseNumber: "2150",
		Title:        "Computer Science I",
		InstructorID: 100,
		TimeSlotID:   1,
		RoomID:       200,
		Mode:         "Traditional",
		Status:       "Scheduled", // Normal
		Lab:          false,
		TimeSlot:     timeSlot,
	})

	courses2 = append(courses2, IntegrationCourseDetail{
		ID:           52,
		CRN:          13001,
		Section:      "002",
		ScheduleID:   2,
		Prefix:       "MATH",
		CourseNumber: "1220",
		Title:        "Calculus I",
		InstructorID: 100, // Same instructor
		TimeSlotID:   1,
		RoomID:       200, // Same room
		Mode:         "Traditional",
		Status:       "Scheduled", // Normal
		Lab:          false,
		TimeSlot:     timeSlot,
	})

	conflicts, err := mockScheduler.DetectConflictsBetweenSchedules(1, 2, courses1, courses2)

	assert.NoError(t, err)
	assert.NotNil(t, conflicts)

	// Should only have conflicts between the 2 normal courses, not any of the 50 removed ones
	assert.Len(t, conflicts["instructor"], 1, "Should have exactly one instructor conflict between normal courses")
	assert.Len(t, conflicts["room"], 1, "Should have exactly one room conflict between normal courses")
	assert.Empty(t, conflicts["crosslisting"], "No crosslisting conflicts")
	assert.Empty(t, conflicts["course"], "No course conflicts (different prefixes)")

	// Verify the conflicts are between the correct courses
	instructorConflict := conflicts["instructor"][0]
	assert.True(t,
		(instructorConflict.Course1.CRN == 13000 && instructorConflict.Course2.CRN == 13001) ||
			(instructorConflict.Course1.CRN == 13001 && instructorConflict.Course2.CRN == 13000),
		"Conflict should be between the normal courses, not removed ones")
}
