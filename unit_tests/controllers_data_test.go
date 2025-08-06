package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test data structures
type TestCourse struct {
	ID           int    `json:"id"`
	Subject      string `json:"subject"`
	Number       string `json:"number"`
	Title        string `json:"title"`
	Credits      int    `json:"credits"`
	Prefix       string `json:"prefix"` // NEW: moved from schedules
	InstructorID int    `json:"instructor_id"`
	RoomID       int    `json:"room_id"`
	TimeslotID   int    `json:"timeslot_id"`
	ScheduleID   int    `json:"schedule_id"` // Added for relationship
	Mode         string `json:"mode"`        // Added mode field
	Active       bool   `json:"active"`
}

type TestRoom struct {
	ID       int    `json:"id"`
	Number   string `json:"number"`
	Building string `json:"building"`
	Capacity int    `json:"capacity"`
}

type TestInstructor struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type TestTimeSlot struct {
	ID        int    `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Days      string `json:"days"`
}

// Test Courses Management
func TestCoursesManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock courses data
	mockCourses := []TestCourse{
		{ID: 1, Subject: "CS", Number: "101", Title: "Intro to Programming", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP", Active: true},
		{ID: 2, Subject: "CS", Number: "201", Title: "Data Structures", Credits: 4, Prefix: "CS", ScheduleID: 1, Mode: "FSO", Active: true},
		{ID: 3, Subject: "MATH", Number: "101", Title: "Calculus I", Credits: 4, Prefix: "MATH", ScheduleID: 2, Mode: "H", Active: false},
	}

	// Mock save courses handler
	saveCoursesHandler := func(c *gin.Context) {
		// Check authentication
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		var courses []TestCourse
		if err := c.ShouldBindJSON(&courses); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Mock validation
		for _, course := range courses {
			if course.Subject == "" || course.Number == "" || course.Title == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Courses updated successfully"})
	}

	// Mock get courses handler
	getCoursesHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.JSON(http.StatusOK, gin.H{"courses": mockCourses})
	}

	router.POST("/courses/save", saveCoursesHandler)
	router.GET("/courses", getCoursesHandler)

	// Test saving courses with authentication
	t.Run("Save Courses - Authenticated", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		coursesJSON, _ := json.Marshal(mockCourses)
		req, _ := http.NewRequest("POST", "/courses/save", strings.NewReader(string(coursesJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Courses updated successfully", response["message"])
	})

	// Test saving courses without authentication
	t.Run("Save Courses - Not Authenticated", func(t *testing.T) {
		coursesJSON, _ := json.Marshal(mockCourses)
		req, _ := http.NewRequest("POST", "/courses/save", strings.NewReader(string(coursesJSON)))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test saving courses with invalid data
	t.Run("Save Courses - Invalid Data", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		invalidCourses := []TestCourse{
			{ID: 1, Subject: "", Number: "101", Title: "Invalid Course", Prefix: "CS", ScheduleID: 1, Mode: "IP"}, // Missing subject
		}

		coursesJSON, _ := json.Marshal(invalidCourses)
		req, _ := http.NewRequest("POST", "/courses/save", strings.NewReader(string(coursesJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test getting courses
	t.Run("Get Courses", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		req, _ := http.NewRequest("GET", "/courses", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "courses")
	})
}

// Test Rooms Management
func TestRoomsManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock save rooms handler
	saveRoomsHandler := func(c *gin.Context) {
		action := c.PostForm("action")

		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		switch action {
		case "save":
			// Mock save logic
			c.JSON(http.StatusOK, gin.H{"message": "Rooms saved successfully"})
		case "delete":
			selectedRooms := c.PostFormArray("selectedRooms")
			if len(selectedRooms) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No rooms selected"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Rooms deleted successfully"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		}
	}

	// Mock add room handler
	addRoomHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		number := c.PostForm("number")
		building := c.PostForm("building")
		capacityStr := c.PostForm("capacity")

		if number == "" || building == "" || capacityStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		capacity, err := strconv.Atoi(capacityStr)
		if err != nil || capacity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid capacity"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Room added successfully"})
	}

	router.POST("/rooms/save", saveRoomsHandler)
	router.POST("/rooms/add", addRoomHandler)

	// Test saving rooms
	t.Run("Save Rooms", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		form := url.Values{}
		form.Add("action", "save")

		req, _ := http.NewRequest("POST", "/rooms/save", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test deleting rooms
	t.Run("Delete Rooms", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		form := url.Values{}
		form.Add("action", "delete")
		form.Add("selectedRooms", "1")
		form.Add("selectedRooms", "2")

		req, _ := http.NewRequest("POST", "/rooms/save", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test adding room
	t.Run("Add Room - Valid Data", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		form := url.Values{}
		form.Add("number", "101")
		form.Add("building", "Science Hall")
		form.Add("capacity", "30")

		req, _ := http.NewRequest("POST", "/rooms/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test adding room with invalid capacity
	t.Run("Add Room - Invalid Capacity", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "testuser")

		form := url.Values{}
		form.Add("number", "102")
		form.Add("building", "Science Hall")
		form.Add("capacity", "invalid")

		req, _ := http.NewRequest("POST", "/rooms/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test User Management
func TestUserManagement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Mock save users handler
	saveUsersHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("username") == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		var users []map[string]interface{}
		if err := c.ShouldBindJSON(&users); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Mock validation
		for _, user := range users {
			username, hasUsername := user["username"].(string)
			email, hasEmail := user["email"].(string)

			if !hasUsername || !hasEmail || username == "" || email == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
				return
			}

			// Check for password fields
			newPassword, hasNewPassword := user["newPassword"].(string)
			confirmPassword, hasConfirmPassword := user["confirmPassword"].(string)

			if hasNewPassword && hasConfirmPassword {
				if newPassword != "" && newPassword != confirmPassword {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
					return
				}
				if newPassword != "" && len(newPassword) < 6 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Users updated successfully"})
	}

	router.POST("/users/save", saveUsersHandler)

	// Test saving users with valid data
	t.Run("Save Users - Valid Data", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		users := []map[string]interface{}{
			{
				"id":              1,
				"username":        "testuser",
				"email":           "test@example.com",
				"newPassword":     "",
				"confirmPassword": "",
			},
		}

		usersJSON, _ := json.Marshal(users)
		req, _ := http.NewRequest("POST", "/users/save", strings.NewReader(string(usersJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test saving users with password change
	t.Run("Save Users - Password Change", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		users := []map[string]interface{}{
			{
				"id":              1,
				"username":        "testuser",
				"email":           "test@example.com",
				"newPassword":     "newpassword123",
				"confirmPassword": "newpassword123",
			},
		}

		usersJSON, _ := json.Marshal(users)
		req, _ := http.NewRequest("POST", "/users/save", strings.NewReader(string(usersJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test saving users with mismatched passwords
	t.Run("Save Users - Password Mismatch", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		users := []map[string]interface{}{
			{
				"id":              1,
				"username":        "testuser",
				"email":           "test@example.com",
				"newPassword":     "password123",
				"confirmPassword": "different123",
			},
		}

		usersJSON, _ := json.Marshal(users)
		req, _ := http.NewRequest("POST", "/users/save", strings.NewReader(string(usersJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test saving users with short password
	t.Run("Save Users - Short Password", func(t *testing.T) {
		cookie := CreateAuthenticatedSession(router, "admin")

		users := []map[string]interface{}{
			{
				"id":              1,
				"username":        "testuser",
				"email":           "test@example.com",
				"newPassword":     "123",
				"confirmPassword": "123",
			},
		}

		usersJSON, _ := json.Marshal(users)
		req, _ := http.NewRequest("POST", "/users/save", strings.NewReader(string(usersJSON)))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Test Generic Data Validation
func TestDataValidation(t *testing.T) {
	t.Run("Course Validation", func(t *testing.T) {
		validCourses := []TestCourse{
			{Subject: "CS", Number: "101", Title: "Programming", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP"},
			{Subject: "MATH", Number: "201", Title: "Calculus", Credits: 4, Prefix: "MATH", ScheduleID: 1, Mode: "FSO"},
		}

		invalidCourses := []TestCourse{
			{Subject: "", Number: "101", Title: "Invalid", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP"},   // Missing subject
			{Subject: "CS", Number: "", Title: "Invalid", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP"},    // Missing number
			{Subject: "CS", Number: "101", Title: "", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP"},        // Missing title
			{Subject: "CS", Number: "101", Title: "Invalid", Credits: 0, Prefix: "CS", ScheduleID: 1, Mode: "IP"}, // Invalid credits
			{Subject: "CS", Number: "101", Title: "Valid", Credits: 3, Prefix: "", ScheduleID: 1, Mode: "IP"},     // Missing prefix
			{Subject: "CS", Number: "101", Title: "Valid", Credits: 3, Prefix: "CS", ScheduleID: 0, Mode: "IP"},   // Invalid schedule ID
			{Subject: "CS", Number: "101", Title: "Valid", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: ""},     // Missing mode
		}

		for _, course := range validCourses {
			assert.NotEmpty(t, course.Subject, "Course subject should not be empty")
			assert.NotEmpty(t, course.Number, "Course number should not be empty")
			assert.NotEmpty(t, course.Title, "Course title should not be empty")
			assert.Greater(t, course.Credits, 0, "Course credits should be positive")
			assert.NotEmpty(t, course.Prefix, "Course prefix should not be empty")
			assert.Greater(t, course.ScheduleID, 0, "Course schedule ID should be positive")
			assert.NotEmpty(t, course.Mode, "Course mode should not be empty")
		}

		for _, course := range invalidCourses {
			isInvalid := course.Subject == "" || course.Number == "" || course.Title == "" ||
				course.Credits <= 0 || course.Prefix == "" || course.ScheduleID <= 0 || course.Mode == ""
			assert.True(t, isInvalid, "Invalid course should be detected")
		}
	})

	t.Run("Room Validation", func(t *testing.T) {
		validRooms := []TestRoom{
			{Number: "101", Building: "Science Hall", Capacity: 30},
			{Number: "A202", Building: "Arts Building", Capacity: 50},
		}

		invalidRooms := []TestRoom{
			{Number: "", Building: "Science Hall", Capacity: 30},   // Missing number
			{Number: "101", Building: "", Capacity: 30},            // Missing building
			{Number: "101", Building: "Science Hall", Capacity: 0}, // Invalid capacity
		}

		for _, room := range validRooms {
			assert.NotEmpty(t, room.Number, "Room number should not be empty")
			assert.NotEmpty(t, room.Building, "Room building should not be empty")
			assert.Greater(t, room.Capacity, 0, "Room capacity should be positive")
		}

		for _, room := range invalidRooms {
			isInvalid := room.Number == "" || room.Building == "" || room.Capacity <= 0
			assert.True(t, isInvalid, "Invalid room should be detected")
		}
	})

	t.Run("Instructor Validation", func(t *testing.T) {
		validInstructors := []TestInstructor{
			{FirstName: "John", LastName: "Doe", Email: "john.doe@university.edu"},
			{FirstName: "Jane", LastName: "Smith", Email: "jane.smith@university.edu"},
		}

		invalidInstructors := []TestInstructor{
			{FirstName: "", LastName: "Doe", Email: "john.doe@university.edu"},  // Missing first name
			{FirstName: "John", LastName: "", Email: "john.doe@university.edu"}, // Missing last name
			{FirstName: "John", LastName: "Doe", Email: ""},                     // Missing email
			{FirstName: "John", LastName: "Doe", Email: "invalid-email"},        // Invalid email
		}

		for _, instructor := range validInstructors {
			assert.NotEmpty(t, instructor.FirstName, "Instructor first name should not be empty")
			assert.NotEmpty(t, instructor.LastName, "Instructor last name should not be empty")
			assert.NotEmpty(t, instructor.Email, "Instructor email should not be empty")
			assert.Contains(t, instructor.Email, "@", "Instructor email should contain @")
		}

		for _, instructor := range invalidInstructors {
			isInvalid := instructor.FirstName == "" || instructor.LastName == "" ||
				instructor.Email == "" || !strings.Contains(instructor.Email, "@")
			assert.True(t, isInvalid, "Invalid instructor should be detected")
		}
	})

	t.Run("Mode Validation", func(t *testing.T) {
		validModes := []string{"IP", "FSO", "PSO", "H", "CLAS", "AO"}
		invalidModes := []string{"", "INVALID", "XYZ", "123"}
		allowedModes := []string{"IP", "FSO", "PSO", "H", "CLAS", "AO"}

		for _, mode := range validModes {
			assert.NotEmpty(t, mode, "Valid mode should not be empty")
			assert.Contains(t, allowedModes, mode, "Mode should be one of the valid options")
		}

		for _, mode := range invalidModes {
			isInvalid := mode == ""
			if mode != "" {
				isInvalid = true
				for _, allowed := range allowedModes {
					if mode == allowed {
						isInvalid = false
						break
					}
				}
			}
			assert.True(t, isInvalid, "Invalid mode should be detected: "+mode)
		}
	})

	t.Run("Course-Prefix Relationship", func(t *testing.T) {
		// Test that courses can have different prefixes within the same schedule
		coursesWithMixedPrefixes := []TestCourse{
			{Subject: "CS", Number: "101", Title: "Programming I", Credits: 3, Prefix: "CS", ScheduleID: 1, Mode: "IP"},
			{Subject: "MATH", Number: "201", Title: "Calculus", Credits: 4, Prefix: "MATH", ScheduleID: 1, Mode: "FSO"},
			{Subject: "ENG", Number: "101", Title: "English Comp", Credits: 3, Prefix: "ENG", ScheduleID: 1, Mode: "H"},
		}

		// Verify that all courses can belong to the same schedule but have different prefixes
		scheduleID := 1
		prefixes := make(map[string]bool)

		for _, course := range coursesWithMixedPrefixes {
			assert.Equal(t, scheduleID, course.ScheduleID, "All courses should belong to the same schedule")
			assert.NotEmpty(t, course.Prefix, "Course prefix should not be empty")
			prefixes[course.Prefix] = true
		}

		// Verify we have multiple different prefixes
		assert.Greater(t, len(prefixes), 1, "Schedule should have courses with multiple different prefixes")
		assert.Contains(t, prefixes, "CS", "Should have CS prefix")
		assert.Contains(t, prefixes, "MATH", "Should have MATH prefix")
		assert.Contains(t, prefixes, "ENG", "Should have ENG prefix")
	})
}
