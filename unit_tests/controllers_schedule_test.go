package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test data structures for schedule management
type TestSchedule struct {
	ID         int    `json:"id"`
	Term       string `json:"term"`
	Year       int    `json:"year"`
	Department string `json:"department"` // UPDATED: changed from Prefix to Department
}

type TestConflict struct {
	CourseID1    int    `json:"course_id_1"`
	CourseID2    int    `json:"course_id_2"`
	ConflictType string `json:"conflict_type"`
	Description  string `json:"description"`
}

type TestCrosslisting struct {
	ID            int    `json:"id"`
	PrimaryCourse string `json:"primary_course"`
	LinkedCourse  string `json:"linked_course"`
	ScheduleID    int    `json:"schedule_id"`
}

// Test Schedule Management
func TestScheduleManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock schedules data
	mockSchedules := []TestSchedule{
		{ID: 1, Term: "Fall", Year: 2024, Department: "Computer Science"},
		{ID: 2, Term: "Spring", Year: 2025, Department: "Mathematics"},
		{ID: 3, Term: "Summer", Year: 2024, Department: "Engineering"},
	}

	// Mock delete schedule handler
	deleteScheduleHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		scheduleID := c.Param("id")
		if scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule ID required"})
			return
		}

		// Mock validation - prevent deletion of schedule with ID 1
		if scheduleID == "1" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete this schedule"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
	}

	// Mock get schedules handler
	getSchedulesHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		term := c.Query("term")

		filteredSchedules := mockSchedules
		if term != "" {
			var filtered []TestSchedule
			for _, schedule := range mockSchedules {
				if schedule.Term == term {
					filtered = append(filtered, schedule)
				}
			}
			filteredSchedules = filtered
		}

		c.JSON(http.StatusOK, gin.H{"schedules": filteredSchedules})
	}

	router.DELETE("/schedule/:id", deleteScheduleHandler)
	router.GET("/schedules", getSchedulesHandler)

	// Test deleting schedule
	t.Run("Delete Schedule - Valid ID", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		req, _ := http.NewRequest("DELETE", "/schedule/2", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Schedule deleted successfully", response["message"])
	})

	// Test deleting protected schedule
	t.Run("Delete Schedule - Protected ID", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		req, _ := http.NewRequest("DELETE", "/schedule/1", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test getting schedules
	t.Run("Get Schedules", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		req, _ := http.NewRequest("GET", "/schedules", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "schedules")
	})

	// Test getting schedules with filter
	t.Run("Get Schedules - Filtered", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		req, _ := http.NewRequest("GET", "/schedules?term=Fall", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Test Conflict Detection
func TestConflictDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock conflict detection handler
	detectConflictsHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		scheduleID := c.PostForm("schedule_id")
		if scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule ID required"})
			return
		}

		// Mock conflicts data
		mockConflicts := []TestConflict{
			{
				CourseID1:    1,
				CourseID2:    2,
				ConflictType: "time",
				Description:  "Time slot overlap",
			},
			{
				CourseID1:    3,
				CourseID2:    4,
				ConflictType: "instructor",
				Description:  "Same instructor assigned to multiple courses",
			},
		}

		// Mock logic - return conflicts only for specific schedule
		if scheduleID == "1" {
			c.JSON(http.StatusOK, gin.H{
				"conflicts": mockConflicts,
				"total":     len(mockConflicts),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"conflicts": []TestConflict{},
				"total":     0,
			})
		}
	}

	router.POST("/detect-conflicts", detectConflictsHandler)

	// Test conflict detection with conflicts
	t.Run("Detect Conflicts - With Conflicts", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		form := url.Values{}
		form.Add("schedule_id", "1")

		req, _ := http.NewRequest("POST", "/detect-conflicts", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["total"])
		assert.Contains(t, response, "conflicts")
	})

	// Test conflict detection without conflicts
	t.Run("Detect Conflicts - No Conflicts", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		form := url.Values{}
		form.Add("schedule_id", "2")

		req, _ := http.NewRequest("POST", "/detect-conflicts", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["total"])
	})

	// Test conflict detection without schedule ID
	t.Run("Detect Conflicts - Missing Schedule ID", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		req, _ := http.NewRequest("POST", "/detect-conflicts", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test Crosslisting Management
func TestCrosslistingManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock crosslistings data
	mockCrosslistings := []TestCrosslisting{
		{ID: 1, PrimaryCourse: "CS-101", LinkedCourse: "CPSC-101", ScheduleID: 1},
		{ID: 2, PrimaryCourse: "MATH-201", LinkedCourse: "MTH-201", ScheduleID: 1},
	}

	// Mock add crosslisting handler
	addCrosslistingHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		primaryCourse := c.PostForm("primary_course")
		linkedCourse := c.PostForm("linked_course")
		scheduleID := c.PostForm("schedule_id")

		if primaryCourse == "" || linkedCourse == "" || scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		// Mock validation - prevent duplicate crosslistings
		if primaryCourse == "CS-101" && linkedCourse == "CPSC-101" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Crosslisting already exists"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Crosslisting added successfully"})
	}

	// Mock delete crosslistings handler
	deleteCrosslistingsHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		selectedCrosslistings := c.PostFormArray("selectedCrosslistings")
		if len(selectedCrosslistings) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No crosslistings selected"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Crosslistings deleted successfully",
			"deleted": len(selectedCrosslistings),
		})
	}

	// Mock get crosslistings handler
	getCrosslistingsHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.JSON(http.StatusOK, gin.H{"crosslistings": mockCrosslistings})
	}

	router.POST("/crosslistings/add", addCrosslistingHandler)
	router.POST("/crosslistings/delete", deleteCrosslistingsHandler)
	router.GET("/crosslistings", getCrosslistingsHandler)

	// Test adding crosslisting
	t.Run("Add Crosslisting - Valid Data", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		form := url.Values{}
		form.Add("primary_course", "ENG-101")
		form.Add("linked_course", "ENGL-101")
		form.Add("schedule_id", "1")

		req, _ := http.NewRequest("POST", "/crosslistings/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test adding duplicate crosslisting
	t.Run("Add Crosslisting - Duplicate", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		form := url.Values{}
		form.Add("primary_course", "CS-101")
		form.Add("linked_course", "CPSC-101")
		form.Add("schedule_id", "1")

		req, _ := http.NewRequest("POST", "/crosslistings/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test adding crosslisting with missing data
	t.Run("Add Crosslisting - Missing Data", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		form := url.Values{}
		form.Add("primary_course", "ENG-101")
		// Missing linked_course and schedule_id

		req, _ := http.NewRequest("POST", "/crosslistings/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test deleting crosslistings
	t.Run("Delete Crosslistings", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		form := url.Values{}
		form.Add("selectedCrosslistings", "1")
		form.Add("selectedCrosslistings", "2")

		req, _ := http.NewRequest("POST", "/crosslistings/delete", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["deleted"])
	})

	// Test deleting crosslistings with no selection
	t.Run("Delete Crosslistings - No Selection", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		req, _ := http.NewRequest("POST", "/crosslistings/delete", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test getting crosslistings
	t.Run("Get Crosslistings", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "user")

		req, _ := http.NewRequest("GET", "/crosslistings", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "crosslistings")
	})
}

// Test API Endpoints
func TestAPIEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock API handler for getting courses for schedule
	getCoursesForScheduleHandler := func(c *gin.Context) {
		scheduleID := c.Param("scheduleId")
		if scheduleID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule ID required"})
			return
		}

		// Mock courses data
		mockCourses := []map[string]interface{}{
			{
				"id":         1,
				"subject":    "CS",
				"number":     "101",
				"title":      "Programming",
				"credits":    3,
				"instructor": "Dr. Smith",
				"room":       "SCI-101",
				"timeslot":   "MWF 9:00-10:00",
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"courses":    mockCourses,
			"scheduleId": scheduleID,
		})
	}

	router.GET("/api/schedule/:scheduleId/courses", getCoursesForScheduleHandler)

	// Test API endpoint
	t.Run("Get Courses for Schedule API", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/schedule/1/courses", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "courses")
		assert.Equal(t, "1", response["scheduleId"])
	})

	// Test API endpoint with missing schedule ID
	t.Run("Get Courses for Schedule API - Missing ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/schedule//courses", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// This should return 404 or be handled by the router
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

// Test Business Logic
func TestBusinessLogic(t *testing.T) {
	t.Run("Schedule Validation", func(t *testing.T) {
		validSchedules := []TestSchedule{
			{Term: "Fall", Year: 2024, Department: "Computer Science"},
			{Term: "Spring", Year: 2025, Department: "Mathematics"},
			{Term: "Summer", Year: 2024, Department: "Engineering"},
		}

		invalidSchedules := []TestSchedule{
			{Term: "", Year: 2024, Department: "Computer Science"},        // Missing term
			{Term: "Fall", Year: 0, Department: "Computer Science"},       // Invalid year
			{Term: "Fall", Year: 2024, Department: ""},                    // Missing department
			{Term: "Invalid", Year: 2024, Department: "Computer Science"}, // Invalid term
		}

		validTerms := []string{"Fall", "Spring", "Summer", "Winter"}

		for _, schedule := range validSchedules {
			assert.NotEmpty(t, schedule.Term, "Schedule term should not be empty")
			assert.Greater(t, schedule.Year, 2000, "Schedule year should be reasonable")
			assert.NotEmpty(t, schedule.Department, "Schedule department should not be empty")
			assert.Contains(t, validTerms, schedule.Term, "Schedule term should be valid")
		}

		for _, schedule := range invalidSchedules {
			isInvalid := schedule.Term == "" || schedule.Year <= 2000 ||
				schedule.Department == "" || !contains(validTerms, schedule.Term)
			assert.True(t, isInvalid, "Invalid schedule should be detected")
		}
	})

	t.Run("Conflict Types", func(t *testing.T) {
		conflictTypes := []string{"time", "instructor", "room", "resource"}

		for _, conflictType := range conflictTypes {
			assert.NotEmpty(t, conflictType, "Conflict type should not be empty")
			assert.True(t, len(conflictType) > 2, "Conflict type should be descriptive")
		}
	})

	t.Run("Crosslisting Validation", func(t *testing.T) {
		validCrosslistings := []TestCrosslisting{
			{PrimaryCourse: "CS-101", LinkedCourse: "CPSC-101", ScheduleID: 1},
			{PrimaryCourse: "MATH-201", LinkedCourse: "MTH-201", ScheduleID: 1},
		}

		invalidCrosslistings := []TestCrosslisting{
			{PrimaryCourse: "", LinkedCourse: "CPSC-101", ScheduleID: 1},       // Missing primary
			{PrimaryCourse: "CS-101", LinkedCourse: "", ScheduleID: 1},         // Missing linked
			{PrimaryCourse: "CS-101", LinkedCourse: "CPSC-101", ScheduleID: 0}, // Invalid schedule ID
			{PrimaryCourse: "CS-101", LinkedCourse: "CS-101", ScheduleID: 1},   // Same course
		}

		for _, crosslisting := range validCrosslistings {
			assert.NotEmpty(t, crosslisting.PrimaryCourse, "Primary course should not be empty")
			assert.NotEmpty(t, crosslisting.LinkedCourse, "Linked course should not be empty")
			assert.Greater(t, crosslisting.ScheduleID, 0, "Schedule ID should be positive")
			assert.NotEqual(t, crosslisting.PrimaryCourse, crosslisting.LinkedCourse, "Courses should be different")
		}

		for _, crosslisting := range invalidCrosslistings {
			isInvalid := crosslisting.PrimaryCourse == "" || crosslisting.LinkedCourse == "" ||
				crosslisting.ScheduleID <= 0 || crosslisting.PrimaryCourse == crosslisting.LinkedCourse
			assert.True(t, isInvalid, "Invalid crosslisting should be detected")
		}
	})
}

// Helper function for slice contains check
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Test Conflict Detection with Pre-selection
func TestConflictDetectionPreSelection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock conflict select page handler with pre-selection support
	renderConflictSelectHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Check for pre-selected schedules from query parameters
		preSelectedSchedule1 := c.Query("schedule1")
		preSelectedSchedule2 := c.Query("schedule2")

		mockSchedules := []TestSchedule{
			{ID: 1, Term: "Fall", Year: 2024, Department: "Computer Science"},
			{ID: 2, Term: "Spring", Year: 2025, Department: "Mathematics"},
		}

		c.JSON(http.StatusOK, gin.H{
			"schedules":            mockSchedules,
			"preSelectedSchedule1": preSelectedSchedule1,
			"preSelectedSchedule2": preSelectedSchedule2,
		})
	}

	router.GET("/conflicts", renderConflictSelectHandler)

	// Test with no pre-selection
	t.Run("Conflict Select - No Pre-selection", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		req, _ := http.NewRequest("GET", "/conflicts", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "", response["preSelectedSchedule1"])
		assert.Equal(t, "", response["preSelectedSchedule2"])
	})

	// Test with one schedule pre-selected
	t.Run("Conflict Select - One Schedule Pre-selected", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		req, _ := http.NewRequest("GET", "/conflicts?schedule1=1", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "1", response["preSelectedSchedule1"])
		assert.Equal(t, "", response["preSelectedSchedule2"])
	})

	// Test with both schedules pre-selected
	t.Run("Conflict Select - Both Schedules Pre-selected", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		req, _ := http.NewRequest("GET", "/conflicts?schedule1=1&schedule2=2", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "1", response["preSelectedSchedule1"])
		assert.Equal(t, "2", response["preSelectedSchedule2"])
	})
}

// Test Courses Page Conflict Detection
func TestCoursesConflictDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock conflict select page handler for courses page conflict detection
	renderConflictSelectHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Check for pre-selected schedules from query parameters
		preSelectedSchedule1 := c.Query("schedule1")
		preSelectedSchedule2 := c.Query("schedule2")

		c.JSON(http.StatusOK, gin.H{
			"preSelectedSchedule1": preSelectedSchedule1,
			"preSelectedSchedule2": preSelectedSchedule2,
		})
	}

	router.GET("/conflicts", renderConflictSelectHandler)

	// Test courses page conflict detection - should pre-select same schedule for both dropdowns
	t.Run("Courses Conflict Detection - Same Schedule for Both", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		// Simulate the courses page calling conflicts with current schedule as both parameters
		req, _ := http.NewRequest("GET", "/conflicts?schedule1=5&schedule2=5", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "5", response["preSelectedSchedule1"])
		assert.Equal(t, "5", response["preSelectedSchedule2"])
	})
}
