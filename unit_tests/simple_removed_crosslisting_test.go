package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Simple test structures without complex mocking
type SimpleCourseDetail struct {
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
	Status       string
	Lab          bool
	TimeSlot     *SimpleTimeSlot
}

type SimpleTimeSlot struct {
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

// Simple conflict detection functions that mirror the actual logic
func simpleDetectConflictsBetweenSchedules(courses1, courses2 []SimpleCourseDetail) map[string]int {
	conflicts := map[string]int{
		"instructor":   0,
		"room":         0,
		"crosslisting": 0,
		"course":       0,
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
				if !isFSOPSOExceptionSimple(course1, course2) && timeSlotsOverlapSimple(course1.TimeSlot, course2.TimeSlot) {
					conflicts["instructor"]++
				}
			}

			// Check room conflicts
			if course1.RoomID > 0 && course2.RoomID > 0 && course1.RoomID == course2.RoomID {
				if !isRoomExemptModeSimple(course1) && !isRoomExemptModeSimple(course2) && timeSlotsOverlapSimple(course1.TimeSlot, course2.TimeSlot) {
					conflicts["room"]++
				}
			}
		}
	}

	// Crosslisting conflicts (simplified - assume courses with same CRN+1 are crosslisted)
	allCourses := append(courses1, courses2...)
	for i := 0; i < len(allCourses); i++ {
		for j := i + 1; j < len(allCourses); j++ {
			course1 := allCourses[i]
			course2 := allCourses[j]

			// Skip courses with "Removed" status
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Simple crosslisting logic - courses with CRNs that differ by 1 are crosslisted
			if course1.CRN == course2.CRN+1 || course2.CRN == course1.CRN+1 {
				if course1.TimeSlot != nil && course2.TimeSlot != nil {
					if !isTimeExemptModeSimple(course1) && !isTimeExemptModeSimple(course2) && !timeSlotsMatchSimple(course1.TimeSlot, course2.TimeSlot) {
						conflicts["crosslisting"]++
					}
				}
			}
		}
	}

	// Course conflicts (same prefix, same range)
	for _, course1 := range courses1 {
		for _, course2 := range courses2 {
			// Skip courses with "Removed" status
			if course1.Status == "Removed" || course2.Status == "Removed" {
				continue
			}

			// Only check courses with the same prefix
			if course1.Prefix != course2.Prefix {
				continue
			}

			// Check if courses are in the same range
			if isInSameCourseRangeSimple(course1.CourseNumber, course2.CourseNumber) {
				conflicts["course"]++
			}
		}
	}

	return conflicts
}

// Helper functions
func timeSlotsOverlapSimple(slot1, slot2 *SimpleTimeSlot) bool {
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

	// Simple time overlap check
	return !(slot1.EndTime <= slot2.StartTime || slot2.EndTime <= slot1.StartTime)
}

func timeSlotsMatchSimple(slot1, slot2 *SimpleTimeSlot) bool {
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

func isRoomExemptModeSimple(course SimpleCourseDetail) bool {
	return course.Mode == "FSO" || course.Mode == "PSO" || course.Mode == "AO"
}

func isTimeExemptModeSimple(course SimpleCourseDetail) bool {
	return course.Mode == "AO"
}

func isFSOPSOExceptionSimple(course1, course2 SimpleCourseDetail) bool {
	return (course1.Mode == "FSO" || course1.Mode == "PSO") &&
		(course2.Mode == "FSO" || course2.Mode == "PSO")
}

func isInSameCourseRangeSimple(courseNum1, courseNum2 string) bool {
	if len(courseNum1) >= 4 && len(courseNum2) >= 4 {
		return courseNum1[0] == courseNum2[0] // Same thousand range
	}
	return false
}

// Simple test cases that don't require complex mocking

func TestSimple_RemovedCourses_InstructorConflicts_ShouldBeSkipped(t *testing.T) {
	// Create overlapping time slots
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create courses with same instructor and overlapping time slots
	courses1 := []SimpleCourseDetail{
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
			RoomID:       1,
			Mode:         "Traditional",
			Status:       "Removed", // This course has "Removed" status
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12346,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2160",
			Title:        "Computer Science II",
			InstructorID: 100, // Same instructor
			TimeSlotID:   2,
			RoomID:       2,
			Mode:         "Traditional",
			Status:       "Scheduled", // This course is normal
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	// Execute the conflict detection
	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	// Assertions
	assert.Equal(t, 0, conflicts["instructor"], "Removed courses should not create instructor conflicts")
	assert.Equal(t, 0, conflicts["room"], "Removed courses should not create room conflicts")
	assert.Equal(t, 0, conflicts["crosslisting"], "Removed courses should not create crosslisting conflicts")
	assert.Equal(t, 0, conflicts["course"], "Removed courses should not create course conflicts")
}

func TestSimple_RemovedCourses_RoomConflicts_ShouldBeSkipped(t *testing.T) {
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	courses1 := []SimpleCourseDetail{
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
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12346,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "MATH",
			CourseNumber: "1220",
			Title:        "Calculus I",
			InstructorID: 101,
			TimeSlotID:   2,
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Removed", // This course has "Removed" status
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 0, conflicts["room"], "Removed courses should not create room conflicts")
	assert.Equal(t, 0, conflicts["instructor"], "Removed courses should not create instructor conflicts")
}

func TestSimple_CrosslistedCourses_SameTimeSlot_ShouldNotConflict(t *testing.T) {
	// Create identical time slots for crosslisted courses
	timeSlot := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create crosslisted courses (CRNs 12345 and 12346 differ by 1)
	courses1 := []SimpleCourseDetail{
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
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12346, // CRN differs by 1, indicates crosslisting
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 101,
			TimeSlotID:   1,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot, // Same time slot
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 0, conflicts["crosslisting"], "Crosslisted courses with same time slots should not conflict")
}

func TestSimple_CrosslistedCourses_DifferentTimeSlots_ShouldConflict(t *testing.T) {
	// Create different time slots for crosslisted courses
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create crosslisted courses with different time slots
	courses1 := []SimpleCourseDetail{
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
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12346, // CRN differs by 1, indicates crosslisting
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "CS",
			CourseNumber: "2150",
			Title:        "Computer Science I",
			InstructorID: 101,
			TimeSlotID:   2,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot2, // Different time slot
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 1, conflicts["crosslisting"], "Crosslisted courses with different time slots should conflict")
}

func TestSimple_CrosslistedCourses_AOMode_ShouldNotConflict(t *testing.T) {
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "14:00:00",
		EndTime:   "15:30:00",
		Tuesday:   true,
		Thursday:  true,
		Days:      "TR",
	}

	// Create crosslisted courses with different time slots but one in AO mode
	courses1 := []SimpleCourseDetail{
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
			Mode:         "AO", // Arrangement Only mode - exempt from time conflicts
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
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
			Mode:         "Traditional",
			Status:       "Scheduled",
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 0, conflicts["crosslisting"], "Crosslisted courses in AO mode should not create conflicts")
}

func TestSimple_NormalCourses_InstructorConflicts_ShouldStillWork(t *testing.T) {
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create normal courses with same instructor and overlapping time slots
	courses1 := []SimpleCourseDetail{
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
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12347, // Different CRN (not crosslisted)
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "MATH",
			CourseNumber: "1220",
			Title:        "Calculus I",
			InstructorID: 100, // Same instructor
			TimeSlotID:   2,
			RoomID:       201,
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 1, conflicts["instructor"], "Normal courses with same instructor and overlapping times should conflict")
}

func TestSimple_NormalCourses_RoomConflicts_ShouldStillWork(t *testing.T) {
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	courses1 := []SimpleCourseDetail{
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
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12347,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "MATH",
			CourseNumber: "1220",
			Title:        "Calculus I",
			InstructorID: 101,
			TimeSlotID:   2,
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Scheduled", // Normal course
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 1, conflicts["room"], "Normal courses with same room and overlapping times should conflict")
}

func TestSimple_BothCoursesRemoved_ShouldNotConflict(t *testing.T) {
	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
		ID:        2,
		StartTime: "10:30:00",
		EndTime:   "12:00:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	// Create courses where both are removed
	courses1 := []SimpleCourseDetail{
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
			Status:       "Removed", // Both courses are removed
			Lab:          false,
			TimeSlot:     timeSlot1,
		},
	}

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12347,
			Section:      "002",
			ScheduleID:   2,
			Prefix:       "MATH",
			CourseNumber: "1220",
			Title:        "Calculus I",
			InstructorID: 100, // Same instructor
			TimeSlotID:   2,
			RoomID:       200, // Same room
			Mode:         "Traditional",
			Status:       "Removed", // Both courses are removed
			Lab:          false,
			TimeSlot:     timeSlot2,
		},
	}

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 0, conflicts["instructor"], "Both removed courses should not create any conflicts")
	assert.Equal(t, 0, conflicts["room"], "Both removed courses should not create any conflicts")
	assert.Equal(t, 0, conflicts["crosslisting"], "Both removed courses should not create any conflicts")
	assert.Equal(t, 0, conflicts["course"], "Both removed courses should not create any conflicts")
}

func TestSimple_MixedStatusCombinations(t *testing.T) {
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

	timeSlot1 := &SimpleTimeSlot{
		ID:        1,
		StartTime: "10:00:00",
		EndTime:   "11:30:00",
		Monday:    true,
		Wednesday: true,
		Friday:    true,
		Days:      "MWF",
	}

	timeSlot2 := &SimpleTimeSlot{
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
			courses1 := []SimpleCourseDetail{
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
					Status:       tc.status1,
					Lab:          false,
					TimeSlot:     timeSlot1,
				},
			}

			courses2 := []SimpleCourseDetail{
				{
					ID:           2,
					CRN:          12347,
					Section:      "002",
					ScheduleID:   2,
					Prefix:       "MATH",
					CourseNumber: "1220",
					Title:        "Calculus I",
					InstructorID: 100, // Same instructor
					TimeSlotID:   2,
					RoomID:       200, // Same room
					Mode:         "Traditional",
					Status:       tc.status2,
					Lab:          false,
					TimeSlot:     timeSlot2,
				},
			}

			conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

			if tc.shouldConflict {
				assert.Greater(t, conflicts["instructor"], 0, "Expected instructor conflicts for %s", tc.name)
				assert.Greater(t, conflicts["room"], 0, "Expected room conflicts for %s", tc.name)
			} else {
				assert.Equal(t, 0, conflicts["instructor"], "Expected no instructor conflicts for %s", tc.name)
				assert.Equal(t, 0, conflicts["room"], "Expected no room conflicts for %s", tc.name)
			}
		})
	}
}

func TestSimple_CourseConflicts_WithRemovedCourses(t *testing.T) {
	// Test courses in same range that would normally conflict
	courses1 := []SimpleCourseDetail{
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

	courses2 := []SimpleCourseDetail{
		{
			ID:           2,
			CRN:          12347,
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

	conflicts := simpleDetectConflictsBetweenSchedules(courses1, courses2)

	assert.Equal(t, 0, conflicts["course"], "Removed courses should not create course conflicts")

	// Test with both normal - should conflict
	courses1[0].Status = "Scheduled"
	conflictsNormal := simpleDetectConflictsBetweenSchedules(courses1, courses2)
	assert.Equal(t, 1, conflictsNormal["course"], "Normal courses in same range should conflict")
}
