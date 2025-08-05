package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"github.com/xuri/excelize/v2"
)

// FormData represents data passed to templates
type FormData struct {
	Error     string
	Success   string
	Values    map[string]string
	CSRFToken string `json:"csrf_token"`
}

// GinFormData represents data passed to templates via Gin
type GinFormData struct {
	Error     string
	Success   string
	Values    map[string]string
	CSRFToken string
}

func (scheduler *wmu_scheduler) ShowSignupForm(w http.ResponseWriter, r *http.Request) {
	scheduler.renderSignupForm(w, r, "", "", nil)
}

func (scheduler *wmu_scheduler) renderSignupForm(w http.ResponseWriter, r *http.Request, errorMsg, successMsg string, values map[string]string) {
	tmplPath := filepath.Join("templates", "signup.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading signup form", http.StatusInternalServerError)
		return
	}

	// For now, use a placeholder token - this should be improved
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		csrfToken = "placeholder-token" // This should be replaced with proper token generation
	}

	data := FormData{
		Error:     errorMsg,
		Success:   successMsg,
		Values:    values,
		CSRFToken: csrfToken,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering signup form", http.StatusInternalServerError)
	}
}

func (scheduler *wmu_scheduler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	// Check for success message from URL parameters
	successMsg := r.URL.Query().Get("success")
	scheduler.renderLoginForm(w, r, "", successMsg, nil)
}

func (scheduler *wmu_scheduler) renderLoginForm(w http.ResponseWriter, r *http.Request, errorMsg, successMsg string, values map[string]string) {
	tmplPath := filepath.Join("templates", "login.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading login form", http.StatusInternalServerError)
		return
	}

	// For now, use a placeholder token - this should be improved
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		csrfToken = "placeholder-token" // This should be replaced with proper token generation
	}

	data := FormData{
		Error:     errorMsg,
		Success:   successMsg,
		Values:    values,
		CSRFToken: csrfToken,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering login form", http.StatusInternalServerError)
	}
}

// Gin-based controller methods for proper CSRF integration
func (scheduler *wmu_scheduler) ShowSignupFormGin(c *gin.Context) {
	// Check for success and error messages from URL parameters
	successMsg := c.Query("success")
	errorMsg := c.Query("error")
	scheduler.renderSignupFormGin(c, errorMsg, successMsg, nil)
}

func (scheduler *wmu_scheduler) renderSignupFormGin(c *gin.Context, errorMsg, successMsg string, values map[string]string) {
	data := GinFormData{
		Error:     errorMsg,
		Success:   successMsg,
		Values:    values,
		CSRFToken: csrf.GetToken(c),
	}

	c.HTML(http.StatusOK, "signup.html", data)
}

func (scheduler *wmu_scheduler) ShowLoginFormGin(c *gin.Context) {
	// Check for success and error messages from URL parameters
	successMsg := c.Query("success")
	errorMsg := c.Query("error")
	scheduler.renderLoginFormGin(c, errorMsg, successMsg, nil)
}

func (scheduler *wmu_scheduler) renderLoginFormGin(c *gin.Context, errorMsg, successMsg string, values map[string]string) {
	data := GinFormData{
		Error:     errorMsg,
		Success:   successMsg,
		Values:    values,
		CSRFToken: csrf.GetToken(c),
	}

	c.HTML(http.StatusOK, "login.html", data)
}

func (scheduler *wmu_scheduler) LoginUserGin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Preserve form values for re-display
	values := map[string]string{
		"username": username,
	}

	if username == "" || password == "" {
		scheduler.renderLoginFormGin(c, "Username and password are required", "", values)
		return
	}

	loggedIn, err := scheduler.AuthenticateUser(username, password)
	if err != nil {
		scheduler.renderLoginFormGin(c, "Error: "+err.Error(), "", values)
		return
	}

	if !loggedIn {
		scheduler.renderLoginFormGin(c, "Invalid username or password", "", values)
		return
	}

	// Set session using Gin sessions middleware
	session := sessions.Default(c)
	session.Set("username", username)
	session.Save()

	err = scheduler.SetUserLoggedInStatus(username, true)
	if err != nil {
		scheduler.renderLoginFormGin(c, "Error updating login status: "+err.Error(), "", values)
		return
	}

	c.Redirect(http.StatusFound, "/scheduler")
}

func (scheduler *wmu_scheduler) SignupUserGin(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	// Preserve form values for re-display
	values := map[string]string{
		"username": username,
		"email":    email,
	}

	if username == "" || email == "" || password == "" || confirmPassword == "" {
		scheduler.renderSignupFormGin(c, "All fields are required", "", values)
		return
	}

	// Check if passwords match
	if password != confirmPassword {
		scheduler.renderSignupFormGin(c, "Passwords do not match. Please try again.", "", values)
		return
	}

	// Use email as username since that's what the database expects
	err := scheduler.AddUser(username, email, password)
	if err != nil {
		// Check for specific database errors
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "username") {
				scheduler.renderSignupFormGin(c, "Username already exists. Please choose a different username.", "", values)
			} else if strings.Contains(err.Error(), "email") {
				scheduler.renderSignupFormGin(c, "Email address already exists. Please use a different email.", "", values)
			} else {
				scheduler.renderSignupFormGin(c, "User already exists with this username or email.", "", values)
			}
		} else {
			scheduler.renderSignupFormGin(c, err.Error(), "", values)
		}
		return
	}

	// Show success message and redirect
	c.Redirect(http.StatusFound, "/scheduler/login?success=Account created successfully")
}

func (scheduler *wmu_scheduler) LogoutUserGin(c *gin.Context) {
	// Get the current user from session before logging out
	session := sessions.Default(c)
	username := session.Get("username")

	if username != nil {
		if usernameStr, ok := username.(string); ok && usernameStr != "" {
			// Update database to mark user as logged out
			err := scheduler.SetUserLoggedInStatus(usernameStr, false)
			if err != nil {
				AppLogger.LogError(fmt.Sprintf("Error updating logout status for user %s", usernameStr), err)
				// Continue with logout even if database update fails
			}
		}
	}

	// Clear the session
	session.Clear()
	err := session.Save()
	if err != nil {
		AppLogger.LogError("Error clearing session", err)
	}

	// Redirect to login page with success message
	c.Redirect(http.StatusFound, "/scheduler/login?success=You have been logged out successfully")
}

func getIntFromInterface(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return i
		}
	}
	return 0
}

func getStringFromInterface(val interface{}) string {
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.Itoa(int(v))
	}
	return ""
}

func (scheduler *wmu_scheduler) RenderCoursesPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Get schedule_id from the URL query parameters
	scheduleID := c.Query("schedule_id")
	if scheduleID == "" {
		scheduleID, err = scheduler.getCurrentSchedule(c)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"Error": "No schedule currently selected. Please select a schedule.",
				"User":  user,
			})
			return
		}
	} else {
		_, err = scheduler.getCurrentSchedule(c)
		if err != nil {
			session := sessions.Default(c)
			session.Set("schedule_id", scheduleID)
			session.Save()
		}
	}

	id, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid schedule_id parameter",
			"User":  user,
		})
		return
	}

	// Fetch courses from the database or service
	courses, err := scheduler.GetActiveCoursesForSchedule(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching courses: " + err.Error(),
			"User":  user,
		})
		return
	}

	scheduleName, err := scheduler.GetScheduleName(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching schedule name: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Fetch additional data needed for dropdowns
	instructors, err := scheduler.GetAllInstructors()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching instructors: " + err.Error(),
			"User":  user,
		})
		return
	}

	rooms, err := scheduler.GetAllRooms()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching rooms: " + err.Error(),
			"User":  user,
		})
		return
	}

	timeSlots, err := scheduler.GetAllTimeSlots()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching time slots: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"User":         user,
		"ScheduleName": scheduleName,
		"ScheduleID":   id,
		"Courses":      courses,
		"Instructors":  instructors,
		"Rooms":        rooms,
		"TimeSlots":    timeSlots,
		"CSRFToken":    csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "courses", data)
}

// SaveCoursesGin handles POST requests to save course changes
func (scheduler *wmu_scheduler) SaveCoursesGin(c *gin.Context) {

	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		AppLogger.LogError("Authentication error in SaveCoursesGin", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	action := c.PostForm("action")
	if action == "export" {
		scheduler.ExportCoursesToExcel(c)
		return
	}

	// Parse the courses JSON data from the form
	coursesJSON := c.PostForm("courses")
	if coursesJSON == "" {
		AppLogger.LogWarning("No courses data provided in form")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No courses data provided"})
		return
	}

	// Parse JSON into course data structures
	var courses []map[string]interface{}
	err = json.Unmarshal([]byte(coursesJSON), &courses)
	if err != nil {
		AppLogger.LogError("Failed to unmarshal courses JSON data", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid courses data format"})
		return
	}

	// Process each course update
	var errors []string
	successCount := 0

	for _, courseData := range courses {
		// Extract course ID
		courseID, ok := courseData["id"].(string)
		if !ok {
			errors = append(errors, "Invalid course ID format")
			continue
		}

		id, err := strconv.Atoi(courseID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid course ID: %s", courseID))
			continue
		}

		// Extract and convert course fields with safe type assertions
		crn := getIntFromInterface(courseData["crn"])
		section := getIntFromInterface(courseData["section"])
		courseNumber := getIntFromInterface(courseData["course_number"])
		title := getStringFromInterface(courseData["title"])

		// Handle credits as min/max if it contains a dash
		var minCredits, maxCredits int
		creditsStr := getStringFromInterface(courseData["credits"])
		if strings.Contains(creditsStr, "-") {
			parts := strings.SplitN(creditsStr, "-", 2)
			minCredits, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			maxCredits, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		} else {
			minCredits, _ = strconv.Atoi(strings.TrimSpace(creditsStr))
			maxCredits = minCredits
		}

		// Handle contact as min/max if it contains a dash
		var minContact, maxContact int
		contactStr := getStringFromInterface(courseData["contact"])
		if strings.Contains(contactStr, "-") {
			parts := strings.SplitN(contactStr, "-", 2)
			minContact, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			maxContact, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		} else {
			minContact, _ = strconv.Atoi(strings.TrimSpace(contactStr))
			maxContact = minContact
		}
		cap := getIntFromInterface(courseData["cap"])
		approval := getIntFromInterface(courseData["approval"])
		lab := getIntFromInterface(courseData["lab"])
		mode := getStringFromInterface(courseData["mode"])
		status := getStringFromInterface(courseData["status"])
		comment := getStringFromInterface(courseData["comment"])

		// Handle nullable foreign keys
		var instructorID = -1
		var timeslotID = -1
		var roomID = -1

		if instructorIDStr := getStringFromInterface(courseData["instructor_id"]); instructorIDStr != "" && instructorIDStr != "<nil>" && instructorIDStr != "null" {
			instructorID = getIntFromInterface(courseData["instructor_id"])
		}

		if timeslotIDStr := getStringFromInterface(courseData["timeslot_id"]); timeslotIDStr != "" && timeslotIDStr != "<nil>" && timeslotIDStr != "null" {
			timeslotID = getIntFromInterface(courseData["timeslot_id"])
		}

		if roomIDStr := getStringFromInterface(courseData["room_id"]); roomIDStr != "" && roomIDStr != "<nil>" && roomIDStr != "null" {
			roomID = getIntFromInterface(courseData["room_id"])
		}

		err = scheduler.AddOrUpdateCourse(crn, section, courseNumber, title, minCredits, maxCredits, minContact, maxContact, cap, approval, lab, instructorID, timeslotID, roomID, mode, status, comment, 0)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to update course ID %d: %v", id, err))
			continue
		}
		successCount++
	}

	// Set session messages and respond
	session := sessions.Default(c)
	if len(errors) > 0 {
		session.Set("error", fmt.Sprintf("%d courses updated, %d errors", successCount, len(errors)))
		session.Save()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%d courses updated, %d errors", successCount, len(errors)),
			"errors":  errors,
		})
	} else {
		session.Set("success", fmt.Sprintf("All %d courses updated successfully", successCount))
		session.Save()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("All %d courses updated successfully", successCount),
		})
	}
}

func (scheduler *wmu_scheduler) RenderAddCoursePageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Get schedule_id from session
	scheduleID, err := scheduler.getCurrentSchedule(c)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "No schedule currently selected. Please select a schedule.",
			"User":  user,
		})
		return
	}

	id, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid schedule_id parameter",
			"User":  user,
		})
		return
	}

	// Get Prefix for the schedule
	prefix, err := scheduler.GetPrefixForSchedule(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching prefix: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get Instructors, Timeslots, and Rooms
	instructors, err := scheduler.GetAllInstructors()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching instructors: " + err.Error(),
			"User":  user,
		})
		return
	}

	timeslots, err := scheduler.GetAllTimeSlots()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching timeslots: " + err.Error(),
			"User":  user,
		})
		return
	}

	rooms, err := scheduler.GetAllRooms()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching rooms: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"User":        user,
		"Prefix":      prefix.Prefix,
		"Instructors": instructors,
		"Timeslots":   timeslots,
		"Rooms":       rooms,
		"CSRFToken":   csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "add_course", data)
}

func (scheduler *wmu_scheduler) AddCourseGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get schedule_id from session
	scheduleID, err := scheduler.getCurrentSchedule(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No schedule selected"})
		return
	}

	scheduleInt, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	// Parse form values
	crn := c.PostForm("crn")
	section := c.PostForm("section")
	courseNumber := c.PostForm("course_number")
	title := c.PostForm("title")
	minCredits := c.PostForm("min_credits")
	maxCredits := c.PostForm("max_credits")
	minContact := c.PostForm("min_contact")
	maxContact := c.PostForm("max_contact")
	cap := c.PostForm("cap")
	approval := c.PostForm("approval")
	lab := c.PostForm("lab")
	instructorID := c.PostForm("instructor_id")
	timeslotID := c.PostForm("timeslot_id")
	roomID := c.PostForm("room_id")
	mode := c.PostForm("mode")
	comment := c.PostForm("comment")

	// Convert to appropriate types
	crnInt, err := strconv.Atoi(crn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CRN"})
		return
	}
	sectionInt, err := strconv.Atoi(section)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid section"})
		return
	}

	courseNumberInt, err := strconv.Atoi(courseNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course number"})
		return
	}

	minCreditsInt, err := strconv.Atoi(minCredits)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min credits"})
		return
	}

	maxCreditsInt, err := strconv.Atoi(maxCredits)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max credits"})
		return
	}

	minContactInt, err := strconv.Atoi(minContact)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min contact"})
		return
	}

	maxContactInt, err := strconv.Atoi(maxContact)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max contact"})
		return
	}

	capInt, err := strconv.Atoi(cap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cap"})
		return
	}

	var approvalInt int
	if approval == "" {
		approvalInt = 0 // Default to 0 if not provided
		err = nil
	} else {
		approvalInt, err = strconv.Atoi(approval)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid approval"})
		return
	}

	var labInt int
	if lab == "" {
		labInt = 0 // Default to 0 if not provided
		err = nil
	} else {
		labInt, err = strconv.Atoi(lab)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lab"})
		return
	}

	var instructorIDInt int
	if instructorID == "" || instructorID == "<nil>" || instructorID == "null" {
		instructorIDInt = -1 // Use -1 for null values
	} else {
		instructorIDInt, err = strconv.Atoi(instructorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instructor ID"})
			return
		}
	}
	var timeslotIDInt int
	if timeslotID == "" || timeslotID == "<nil>" || timeslotID == "null" {
		timeslotIDInt = -1 // Use -1 for null values
	} else {
		timeslotIDInt, err = strconv.Atoi(timeslotID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timeslot ID"})
			return
		}
	}
	var roomIDInt int
	if roomID == "" || roomID == "<nil>" || roomID == "null" {
		roomIDInt = -1 // Use -1 for null values
	} else {
		roomIDInt, err = strconv.Atoi(roomID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
			return
		}
	}

	err = scheduler.AddCourse(
		crnInt, sectionInt, courseNumberInt, title,
		minCreditsInt, maxCreditsInt, minContactInt, maxContactInt,
		capInt, approvalInt == 1, labInt == 1, instructorIDInt, timeslotIDInt,
		roomIDInt, mode, comment, scheduleInt,
	)
	if err != nil {
		// If this is an AJAX request, return JSON error
		if c.GetHeader("Content-Type") == "application/json" || c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			// Set session error message
			session := sessions.Default(c)
			session.Set("error", err.Error())
			session.Save()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Otherwise, set session error and redirect
		session := sessions.Default(c)
		session.Set("error", err.Error())
		session.Save()
		c.Redirect(http.StatusFound, fmt.Sprintf("/scheduler/courses?schedule_id=%d", scheduleInt))
		return
	}

	// Check if this is an AJAX request
	if c.GetHeader("Content-Type") == "application/json" || c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
		// Set session success message
		session := sessions.Default(c)
		session.Set("success", "Course added successfully")
		session.Save()

		// Return JSON for AJAX requests
		courses, err := scheduler.GetActiveCoursesForSchedule(scheduleInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Course added successfully",
			"courses": courses,
		})
	} else {
		// Set session success message and redirect
		session := sessions.Default(c)
		session.Set("success", "Course added successfully")
		session.Save()
		c.Redirect(http.StatusFound, fmt.Sprintf("/scheduler/courses?schedule_id=%d", scheduleInt))
	}
}

func (scheduler *wmu_scheduler) RenderHomePageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Fetch schedule data
	schedules, err := scheduler.GetAllSchedules()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "home.html", gin.H{
			"Error": "Error loading schedules: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Schedules": schedules,
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "home.html", data)
}

// ShowImportPage displays the Excel import page
func (scheduler *wmu_scheduler) ShowImportPage(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get current user
	currentUser, err := scheduler.GetUserByUsername(user.Username)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if currentUser == nil || !currentUser.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  currentUser,
		})
		return
	}

	data := gin.H{
		"User":      currentUser,
		"CSRFToken": csrf.GetToken(c),
	}

	prefixes, err := scheduler.GetAllPrefixes()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching departments: " + err.Error(),
			"User":  currentUser,
		})
		return
	}
	data["Prefixes"] = prefixes

	c.HTML(http.StatusOK, "import.html", data)
}

// Helper function to get current user from session
func (scheduler *wmu_scheduler) getCurrentUser(c *gin.Context) (*User, error) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return nil, fmt.Errorf("no session")
	}

	usernameStr, ok := username.(string)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}

	if usernameStr == "" {
		return nil, fmt.Errorf("empty username in session")
	}

	user, err := scheduler.GetUserByUsername(usernameStr)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (scheduler *wmu_scheduler) getCurrentSchedule(c *gin.Context) (string, error) {
	session := sessions.Default(c)
	scheduleID := session.Get("schedule_id")
	if scheduleID == nil {
		return "", fmt.Errorf("no schedule_id in session")
	}

	return scheduleID.(string), nil
}

// RenderRoomsPageGin renders the rooms page
func (scheduler *wmu_scheduler) RenderRoomsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	rooms, err := scheduler.GetAllRooms()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "rooms", gin.H{
			"Error": "Error loading rooms: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Rooms":     rooms,
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "rooms", data)
}

// RenderTimeslotsPageGin renders the timeslots page
func (scheduler *wmu_scheduler) RenderTimeslotsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Get all time slots
	timeslots, err := scheduler.GetAllTimeSlots()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "timeslots", gin.H{
			"Error": "Error loading timeslots: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"TimeSlots": timeslots,
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "timeslots", data)
}

// RenderInstructorsPageGin renders the instructors page
func (scheduler *wmu_scheduler) RenderInstructorsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	instructors, err := scheduler.GetAllInstructors()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "instructors", gin.H{
			"Error": "Error loading instructors: " + err.Error(),
			"User":  user,
		})
		return
	}

	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "instructors", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get success/error messages from session
	session := sessions.Default(c)
	successMessage := session.Get("success")
	errorMessage := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	data := gin.H{
		"Instructors": instructors,
		"Departments": departments,
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
		"Success":     successMessage,
		"Error":       errorMessage,
	}

	c.HTML(http.StatusOK, "instructors", data)
}

// RenderDepartmentsPageGin renders the departments page
func (scheduler *wmu_scheduler) RenderDepartmentsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "departments.html", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Departments": departments,
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "departments", data)
}

// RenderPrefixesPageGin renders the prefixes page
func (scheduler *wmu_scheduler) RenderPrefixesPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	prefixes, err := scheduler.GetAllPrefixes()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "prefixes", gin.H{
			"Error": "Error loading prefixes: " + err.Error(),
			"User":  user,
		})
		return
	}

	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "prefixes", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Prefixes":    prefixes,
		"Departments": departments,
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "prefixes", data)
}

// RenderUsersPageGin renders the users page
func (scheduler *wmu_scheduler) RenderUsersPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	users, err := scheduler.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "users", gin.H{
			"Error": "Error loading users: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Users":     users,
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if successMsg != nil {
		data["Success"] = successMsg
	}
	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "users", data)
}

func (scheduler *wmu_scheduler) DeleteScheduleGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	session := sessions.Default(c)

	// Get schedule ID from form data
	scheduleID := c.PostForm("schedule_id")
	if scheduleID == "" {
		session.Set("error", "No schedule selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler")
		return
	}

	// Convert scheduleID to integer
	id, err := strconv.Atoi(scheduleID)
	if err != nil {
		session.Set("error", "Invalid schedule ID")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler")
		return
	}

	// Attempt to delete the schedule using the database method
	err = scheduler.DeleteSchedule(id)
	if err != nil {
		session.Set("error", "Failed to delete schedule: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler")
		return
	}

	// Set success message and redirect
	session.Set("success", "Schedule deleted successfully")
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler")
}

// ExcelCourseData represents a course row from Excel
type ExcelCourseData struct {
	CRN               string
	CourseID          string
	Section           string
	Status            string
	Title             string
	Link1             string
	Link2             string
	SchedType         string
	Reserved          string
	MinCreditHours    string
	MaxCreditHours    string
	BillingHours      string
	MinContactHours   string
	MaxContactHours   string
	Gradeable         string
	Capacity          string
	WaitlistCap       string
	SpecialApproval   string
	MeetingType       string
	MeetingTypeDesc   string
	Dates             string
	Days              string
	Time              string
	Location          string
	SiteCode          string
	PrimaryInstructor string
	Fee               string
	Comment           string
}

// ImportExcelSchedule imports course data from Excel file
func (scheduler *wmu_scheduler) ImportExcelSchedule(filePath string, schedule *Schedule) error {

	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("error opening Excel file: %v", err)
	}
	defer f.Close()

	// Get the first sheet (CS)
	sheetName := f.GetSheetList()[0]

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("error reading Excel sheet: %v", err)
	}

	if len(rows) < 6 {
		return fmt.Errorf("insufficient data in Excel file")
	}

	// Headers are in row 5 (index 4)
	headers := rows[4]

	// Create a map of column indices
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.TrimSpace(header)] = i
	}

	// Import courses starting from row 6 (index 5)
	var importedCount int
	var errorCount int

	for i := 5; i < len(rows); i++ {
		row := rows[i]

		// Skip empty rows
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		// Parse course data
		courseData := parseExcelRow(row, columnMap)

		// Skip rows that don't have CRN (likely comment rows)
		if courseData.CRN == "" || !isValidCRN(courseData.CRN) {
			continue
		}

		// Import the course
		err := scheduler.importCourseFromExcel(courseData, schedule)
		if err != nil {
			AppLogger.LogError(fmt.Sprintf("Error importing course CRN %s", courseData.CRN), err)
			errorCount++
		} else {
			importedCount++
		}
	}

	AppLogger.LogInfo(fmt.Sprintf("Import completed: %d courses imported, %d errors", importedCount, errorCount))
	return nil
}

// parseExcelRow parses a row from Excel into ExcelCourseData
func parseExcelRow(row []string, columnMap map[string]int) ExcelCourseData {
	data := ExcelCourseData{}

	// Helper function to get value from column
	getValue := func(columnName string) string {
		if idx, exists := columnMap[columnName]; exists && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	data.CRN = getValue("CRN")
	data.CourseID = getValue("Course ID")
	data.Section = getValue("Section")
	data.Status = getValue("Status")
	data.Title = getValue("Title")
	data.Link1 = getValue("Link1")
	data.Link2 = getValue("Link2")
	data.SchedType = getValue("Sched Type")
	data.Reserved = getValue("Rsvrd")
	creditRange := getValue("Credit Hours")
	if strings.Contains(creditRange, "-") {
		parts := strings.Split(creditRange, "-")
		if len(parts) == 2 {
			data.MinCreditHours = strings.TrimSpace(parts[0])
			data.MaxCreditHours = strings.TrimSpace(parts[1])
		} else {
			data.MinCreditHours = creditRange
			data.MaxCreditHours = creditRange
		}
	} else {
		data.MinCreditHours = creditRange
		data.MaxCreditHours = creditRange
	}
	data.BillingHours = getValue("Billing Hours")
	contactRange := getValue("Contact Hours")
	if strings.Contains(contactRange, "-") {
		parts := strings.Split(contactRange, "-")
		if len(parts) == 2 {
			data.MinContactHours = strings.TrimSpace(parts[0])
			data.MaxContactHours = strings.TrimSpace(parts[1])
		} else {
			data.MinContactHours = contactRange
			data.MaxContactHours = contactRange
		}
	} else {
		data.MinContactHours = contactRange
		data.MaxContactHours = contactRange
	}
	data.Gradeable = getValue("Grad- able")
	data.Capacity = getValue("Cap")
	data.WaitlistCap = getValue("Waitlist Cap")
	data.SpecialApproval = getValue("Spec Appr")
	data.MeetingType = getValue("Mtg Type")
	data.MeetingTypeDesc = getValue("Meeting Type Desc")
	data.Dates = getValue("Dates")
	data.Days = getValue("Days")
	data.Time = getValue("Time")
	data.Location = getValue("Location")
	data.SiteCode = getValue("Site Code")
	data.PrimaryInstructor = getValue("Primary Instructor")
	data.Fee = getValue("Fee")
	data.Comment = getValue("Comment ")

	return data
}

// importCourseFromExcel imports a single course from Excel data
func (scheduler *wmu_scheduler) importCourseFromExcel(data ExcelCourseData, schedule *Schedule) error {
	// Parse course number and prefix from Course ID (e.g., "CS 1110")
	courseParts := strings.Fields(data.CourseID)
	if len(courseParts) < 2 {
		return fmt.Errorf("invalid course ID format: %s", data.CourseID)
	}
	// Check for duplicate schedule
	courseNum := 0
	courseNum, err := strconv.Atoi(courseParts[1])
	if err != nil {
		return fmt.Errorf("invalid course number in Course ID: %s", data.CourseID)
	}

	// Parse CRN
	crn, err := strconv.Atoi(data.CRN)
	if err != nil {
		return fmt.Errorf("invalid CRN: %s", data.CRN)
	}

	// Parse section
	section := data.Section

	// Parse credits
	minCredits, err := strconv.Atoi(data.MinCreditHours)
	if err != nil || minCredits < 0 {
		return fmt.Errorf("invalid credit hours: %s", data.MinCreditHours)
	}

	maxCredits, err := strconv.Atoi(data.MaxCreditHours)
	if err != nil || maxCredits < 0 {
		return fmt.Errorf("invalid credit hours: %s", data.MaxCreditHours)
	}

	// Parse contact hours
	minContactHours, err := strconv.Atoi(data.MinContactHours)
	if err != nil || minContactHours < 0 {
		return fmt.Errorf("invalid contact hours: %s", data.MinContactHours)
	}

	maxContactHours, err := strconv.Atoi(data.MaxContactHours)
	if err != nil || maxContactHours < 0 {
		return fmt.Errorf("invalid contact hours: %s", data.MaxContactHours)
	}

	// Parse capacity
	capacity, err := strconv.Atoi(data.Capacity)
	if err != nil || capacity < 0 {
		return fmt.Errorf("invalid capacity: %s", data.Capacity)
	}

	// Parse time slot
	timeSlotID := -1
	if data.Time != "" && data.Days != "" {
		var err error
		timeSlotID, err = scheduler.findOrCreateTimeSlot(data.Days, data.Time)
		if err != nil {
			AppLogger.LogWarning(fmt.Sprintf("Could not create time slot for %s %s: %v", data.Days, data.Time, err))
			timeSlotID = -1 // This will be converted to NULL
		}
	}

	// Parse room
	roomID := -1
	if data.Location != "" {
		var err error
		roomID, err = scheduler.findOrCreateRoom(data.Location)
		if err != nil {
			AppLogger.LogWarning(fmt.Sprintf("Could not create room for %s: %v", data.Location, err))
			roomID = -1 // This will be converted to NULL
		}
	}

	// Parse instructor
	instructorID := -1
	if data.PrimaryInstructor != "" {
		var err error
		instructorID, err = scheduler.findOrCreateInstructor(data.PrimaryInstructor)
		if err != nil {
			AppLogger.LogWarning(fmt.Sprintf("Could not create instructor for %s: %v", data.PrimaryInstructor, err))
			instructorID = -1 // This will be converted to NULL
		}
	}

	// Parse section as int
	sectionInt, err := strconv.Atoi(section)
	if err != nil {
		return fmt.Errorf("invalid section: %s", section)
	}

	appr := 0
	if strings.TrimSpace(data.SpecialApproval) != "" {
		appr = 1
	}

	lab := 0
	if data.Link1 == "B1" && minCredits == 0 {
		lab = 1
	}

	err = scheduler.AddOrUpdateCourse(crn, sectionInt, courseNum, data.Title,
		minCredits, maxCredits, minContactHours, maxContactHours, capacity, appr, lab, instructorID, timeSlotID,
		roomID, data.MeetingType, "Scheduled", data.Comment, schedule.ID)

	return err
}

// Helper functions
func isValidCRN(crn string) bool {
	return len(crn) == 5 && isNumeric(crn)
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func parseTime(timeStr string) (string, error) {
	// Convert "1130" to "11:30:00"
	if len(timeStr) != 4 {
		return "", fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour := timeStr[:2]
	minute := timeStr[2:]
	return fmt.Sprintf("%s:%s:00", hour, minute), nil
}

// Web handler for Excel import
func (scheduler *wmu_scheduler) ImportExcelHandler(c *gin.Context) {
	// Handle file upload
	file, err := c.FormFile("excel_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Get form parameters
	term := c.PostForm("term")
	yearStr := c.PostForm("year")
	prefixName := c.PostForm("prefix")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	// Save uploaded file
	uploadPath := fmt.Sprintf("uploads/%s", file.Filename)
	err = c.SaveUploadedFile(file, uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create schedule if it doesn't exist, otherwise get existing schedule
	schedule, err := scheduler.AddOrGetSchedule(term, year, prefixName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	session := sessions.Default(c)
	session.Set("schedule_id", strconv.Itoa(schedule.ID))
	session.Save()

	// Import the Excel file
	err = scheduler.ImportExcelSchedule(uploadPath, schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Excel schedule imported successfully!",
		"schedule_id": schedule.ID,
		"redirect":    fmt.Sprintf("/scheduler/courses?schedule_id=%d", schedule.ID),
	})
}

// UpdateCourseGin handles AJAX PUT requests to update a course field
func (scheduler *wmu_scheduler) UpdateCourseGin(c *gin.Context) {
	var req struct {
		CourseID int    `json:"course_id"`
		Field    string `json:"field"`
		Value    string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate field name (add allowed fields as needed)
	allowedFields := map[string]bool{
		"crn":           true,
		"section":       true,
		"course_number": true,
		"title":         true,
		"credits":       true,
		"contacts":      true,
		"cap":           true,
		"approval":      true,
		"lab":           true,
		"instructor_id": true,
		"room_id":       true,
		"time_slot_id":  true,
		"mode":          true,
		"status":        true,
		"comment":       true,
		// Add other fields as needed
	}

	if !allowedFields[req.Field] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid field"})
		return
	}

	if req.Field == "credits" {
		if strings.Contains(req.Value, "-") {
			parts := strings.Split(req.Value, "-")
			if len(parts) == 2 {
				minCredits := strings.TrimSpace(parts[0])
				maxCredits := strings.TrimSpace(parts[1])
				if err := scheduler.UpdateCourseField(req.CourseID, "min_credits", minCredits); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				if err := scheduler.UpdateCourseField(req.CourseID, "max_credits", maxCredits); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Course credits updated successfully"})
				return
			}
		}
	}

	if req.Field == "contact" {
		if strings.Contains(req.Value, "-") {
			parts := strings.Split(req.Value, "-")
			if len(parts) == 2 {
				minContact := strings.TrimSpace(parts[0])
				maxContact := strings.TrimSpace(parts[1])
				if err := scheduler.UpdateCourseField(req.CourseID, "min_contact", minContact); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				if err := scheduler.UpdateCourseField(req.CourseID, "max_contact", maxContact); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Course contact hours updated successfully"})
				return
			}
		}
	}

	// Update the course in the database
	err := scheduler.UpdateCourseField(req.CourseID, req.Field, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course updated successfully"})
}

// RenderAddRoomPageGin renders the add room page
func (scheduler *wmu_scheduler) RenderAddRoomPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	session.Delete("error")
	session.Save()

	data := gin.H{
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "add_room", data)
}

// SaveOrDeleteRoomsGin handles POST requests to save or delete rooms
func (scheduler *wmu_scheduler) SaveOrDeleteRoomsGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle room deletion
		roomIDs := c.PostFormArray("room_ids[]")
		if len(roomIDs) == 0 {
			session := sessions.Default(c)
			session.Set("error", "No rooms selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/rooms")
			return
		}

		var errors []string
		deletedCount := 0

		for _, roomIDStr := range roomIDs {
			roomID, err := strconv.Atoi(roomIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid room ID: %s", roomIDStr))
				continue
			}

			err = scheduler.DeleteRoom(roomID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete room ID %d: %v", roomID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session := sessions.Default(c)
			session.Set("error", fmt.Sprintf("%d rooms deleted, %d errors occurred", deletedCount, len(errors)))
			session.Save()
		} else {
			session := sessions.Default(c)
			session.Set("success", fmt.Sprintf("%d rooms deleted successfully", deletedCount))
			session.Save()
		}
		c.Redirect(http.StatusFound, "/scheduler/rooms")
		return
	}

	// Handle room updates (save operation)
	var successCount, errorCount int

	// Get all room form data
	// The form sends data as rooms[0][ID], rooms[0][Building], etc.
	roomsData := make(map[int]map[string]string)

	// Parse all form values
	for key, values := range c.Request.PostForm {
		if strings.HasPrefix(key, "rooms[") && len(values) > 0 {
			// Extract index and field name from key like "rooms[0][Building]"
			parts := strings.Split(key, "][")
			if len(parts) == 2 {
				indexStr := strings.TrimPrefix(parts[0], "rooms[")
				fieldName := strings.TrimSuffix(parts[1], "]")

				index, err := strconv.Atoi(indexStr)
				if err != nil {
					continue
				}

				if roomsData[index] == nil {
					roomsData[index] = make(map[string]string)
				}
				roomsData[index][fieldName] = values[0]
			}
		}
	}

	// Process each room update
	for _, roomData := range roomsData {
		roomID, err := strconv.Atoi(roomData["ID"])
		if err != nil {
			errorCount++
			continue
		}

		building := roomData["Building"]
		roomNumber := roomData["RoomNumber"]

		capacity, err := strconv.Atoi(roomData["Capacity"])
		if err != nil {
			errorCount++
			continue
		}

		computerLab := roomData["ComputerLab"] == "on"
		dedicatedLab := roomData["DedicatedLab"] == "on"

		err = scheduler.UpdateRoom(roomID, building, roomNumber, capacity, computerLab, dedicatedLab)
		if err != nil {
			errorCount++
			continue
		}
		successCount++
	}

	if errorCount > 0 {
		session := sessions.Default(c)
		session.Set("error", fmt.Sprintf("%d rooms updated, %d errors occurred", successCount, errorCount))
		session.Save()
	} else {
		session := sessions.Default(c)
		session.Set("success", fmt.Sprintf("%d rooms updated successfully", successCount))
		session.Save()
	}
	c.Redirect(http.StatusFound, "/scheduler/rooms")
}

// SaveTimeslotsGin handles POST requests to save timeslot changes and bulk deletion
func (scheduler *wmu_scheduler) SaveTimeslotsGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	session := sessions.Default(c)

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle timeslot deletion
		timeslotIDs := c.PostFormArray("timeslot_ids[]")
		if len(timeslotIDs) == 0 {
			session.Set("error", "No timeslots selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/timeslots")
			return
		}

		var errors []string
		deletedCount := 0

		for _, timeslotIDStr := range timeslotIDs {
			timeslotID, err := strconv.Atoi(timeslotIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid timeslot ID: %s", timeslotIDStr))
				continue
			}

			err = scheduler.DeleteTimeslot(timeslotID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete timeslot ID %d: %v", timeslotID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session.Set("error", fmt.Sprintf("%d timeslots deleted, %d errors occurred", deletedCount, len(errors)))
		} else {
			session.Set("success", fmt.Sprintf("%d timeslots deleted successfully", deletedCount))
		}
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/timeslots")
		return
	}

	// Check for special cases
	if c.PostForm("no_changes") == "true" {
		session.Set("error", "No changes to save")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/timeslots")
		return
	}

	if c.PostForm("no_selection") == "true" {
		session.Set("error", "No timeslots selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/timeslots")
		return
	}

	// Get the timeslots JSON data
	timeslotsJSON := c.PostForm("timeslots")
	if timeslotsJSON == "" {
		session.Set("error", "No timeslot data provided")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/timeslots")
		return
	}

	// Parse the JSON data
	var timeslots []map[string]interface{}
	err = json.Unmarshal([]byte(timeslotsJSON), &timeslots)
	if err != nil {
		session.Set("error", "Invalid timeslot data format")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/timeslots")
		return
	}

	// Update each timeslot
	successCount := 0
	errorCount := 0

	for _, timeslotData := range timeslots {
		id := getIntFromInterface(timeslotData["id"])
		startTime := getStringFromInterface(timeslotData["startTime"])
		endTime := getStringFromInterface(timeslotData["endTime"])
		days := getStringFromInterface(timeslotData["days"])

		if id <= 0 || startTime == "" || endTime == "" || days == "" {
			errorCount++
			continue
		}

		err = scheduler.UpdateTimeslot(id, startTime, endTime, days)
		if err != nil {
			errorCount++
			continue
		}
		successCount++
	}

	// Set appropriate success/error message
	if errorCount > 0 {
		session.Set("error", fmt.Sprintf("%d timeslots updated, %d errors occurred", successCount, errorCount))
	} else {
		session.Set("success", fmt.Sprintf("%d timeslots updated successfully", successCount))
	}
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/timeslots")
}

// RenderAddTimeslotPageGin renders the add timeslot page
func (scheduler *wmu_scheduler) RenderAddTimeslotPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	session.Delete("error")
	session.Save()

	data := gin.H{
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "add_timeslot", data)
}

// AddTimeslotGin handles POST requests to add a new timeslot
func (scheduler *wmu_scheduler) AddTimeslotGin(c *gin.Context) {
	// Get form values
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	Monday := c.PostForm("M")
	Tuesday := c.PostForm("T")
	Wednesday := c.PostForm("W")
	Thursday := c.PostForm("R")
	Friday := c.PostForm("F")

	// Validate required fields
	if startTime == "" || endTime == "" || (Monday == "" && Tuesday == "" && Wednesday == "" && Thursday == "" && Friday == "") {
		session := sessions.Default(c)
		session.Set("error", "All fields are required")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_timeslot")
		return
	}

	// Add the timeslot to the database
	err := scheduler.AddTimeslotWithDays(startTime, endTime, Monday != "", Tuesday != "", Wednesday != "", Thursday != "", Friday != "")
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Failed to add timeslot: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_timeslot")
		return
	}

	// Success - redirect to timeslots page with success message
	session := sessions.Default(c)
	session.Set("success", "Timeslot added successfully")
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/timeslots")
}

// SaveInstructorsGin handles saving instructor changes and bulk deletion
func (scheduler *wmu_scheduler) SaveInstructorsGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	session := sessions.Default(c)

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle instructor deletion
		instructorIDs := c.PostFormArray("instructor_ids[]")
		if len(instructorIDs) == 0 {
			session.Set("error", "No instructors selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/instructors")
			return
		}

		var errors []string
		deletedCount := 0

		for _, instructorIDStr := range instructorIDs {
			instructorID, err := strconv.Atoi(instructorIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid instructor ID: %s", instructorIDStr))
				continue
			}

			err = scheduler.DeleteInstructor(instructorID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete instructor ID %d: %v", instructorID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session.Set("error", fmt.Sprintf("%d instructors deleted, %d errors occurred", deletedCount, len(errors)))
		} else {
			session.Set("success", fmt.Sprintf("%d instructors deleted successfully", deletedCount))
		}
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/instructors")
		return
	}

	// Check for special cases
	if c.PostForm("no_changes") == "true" {
		session.Set("error", "No changes to save")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/instructors")
		return
	}

	if c.PostForm("no_selection") == "true" {
		session.Set("error", "No instructors selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/instructors")
		return
	}

	// Parse the instructors JSON data
	instructorsJSON := c.PostForm("instructors")
	if instructorsJSON == "" {
		session.Set("error", "No instructor data provided")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/instructors")
		return
	}

	var instructors []struct {
		ID         string `json:"id"`
		LastName   string `json:"lastName"`
		FirstName  string `json:"firstName"`
		Department string `json:"department"`
		Status     string `json:"status"`
	}

	err = json.Unmarshal([]byte(instructorsJSON), &instructors)
	if err != nil {
		session.Set("error", "Invalid instructor data format")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/instructors")
		return
	}

	// Update each instructor
	updatedCount := 0
	for _, instructor := range instructors {
		instructorID, err := strconv.Atoi(instructor.ID)
		if err != nil {
			continue // Skip invalid IDs
		}

		err = scheduler.UpdateInstructor(instructorID, instructor.LastName, instructor.FirstName, instructor.Department, instructor.Status)
		if err != nil {
			session.Set("error", fmt.Sprintf("Error updating instructor %s %s: %v", instructor.FirstName, instructor.LastName, err))
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/instructors")
			return
		}
		updatedCount++
	}

	session.Set("success", fmt.Sprintf("Successfully updated %d instructor(s)", updatedCount))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/instructors")
}

// RenderAddInstructorPageGin renders the add instructor page
func (scheduler *wmu_scheduler) RenderAddInstructorPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "add_instructor", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get error message from session
	session := sessions.Default(c)
	errorMessage := session.Get("error")
	session.Delete("error")
	session.Save()

	data := gin.H{
		"Departments": departments,
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
		"Error":       errorMessage,
	}

	c.HTML(http.StatusOK, "add_instructor", data)
}

// AddInstructorGin handles adding a new instructor
func (scheduler *wmu_scheduler) AddInstructorGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	session := sessions.Default(c)

	// Get form data
	lastName := c.PostForm("last_name")
	firstName := c.PostForm("first_name")
	department := c.PostForm("department")
	status := c.PostForm("status")

	// Validate required fields
	if lastName == "" || firstName == "" || department == "" || status == "" {
		session.Set("error", "All fields are required")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_instructor")
		return
	}

	// Add the instructor
	err = scheduler.AddInstructor(lastName, firstName, department, status)
	if err != nil {
		session.Set("error", "Error adding instructor: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_instructor")
		return
	}

	session.Set("success", fmt.Sprintf("Instructor %s %s added successfully", firstName, lastName))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/instructors")
}

// RenderAddUserPageGin renders the add user page
func (scheduler *wmu_scheduler) RenderAddUserPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	scheduler.renderAddUserFormGin(c, "", "", nil)
}

// renderAddUserFormGin renders the add user form with optional error/success messages and preserved values
func (scheduler *wmu_scheduler) renderAddUserFormGin(c *gin.Context, errorMsg, successMsg string, values map[string]string) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session if not provided directly
	session := sessions.Default(c)
	if errorMsg == "" {
		if sessionError := session.Get("error"); sessionError != nil {
			errorMsg = sessionError.(string)
			session.Delete("error")
		}
	}
	if successMsg == "" {
		if sessionSuccess := session.Get("success"); sessionSuccess != nil {
			successMsg = sessionSuccess.(string)
			session.Delete("success")
		}
	}
	session.Save()

	data := gin.H{
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
		"Values":    values,
	}

	if successMsg != "" {
		data["Success"] = successMsg
	}
	if errorMsg != "" {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "add_user", data)
}

// AddUserGin handles adding a new user
func (scheduler *wmu_scheduler) AddUserGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	administrator := c.PostForm("administrator") == "true"

	// Preserve form values for re-display on error
	values := map[string]string{
		"username": username,
		"email":    email,
	}
	if administrator {
		values["administrator"] = "true"
	}

	// Validate required fields
	if username == "" || email == "" || password == "" || confirmPassword == "" {
		scheduler.renderAddUserFormGin(c, "All fields are required", "", values)
		return
	}

	// Check if passwords match
	if password != confirmPassword {
		scheduler.renderAddUserFormGin(c, "Passwords do not match. Please try again.", "", values)
		return
	}

	// Add the user
	err = scheduler.AddUser(username, email, password)
	if err != nil {
		// Check for specific database errors
		if strings.Contains(err.Error(), "Duplicate entry") {
			if strings.Contains(err.Error(), "username") {
				scheduler.renderAddUserFormGin(c, "Username already exists. Please choose a different username.", "", values)
			} else if strings.Contains(err.Error(), "email") {
				scheduler.renderAddUserFormGin(c, "Email address already exists. Please use a different email.", "", values)
			} else {
				scheduler.renderAddUserFormGin(c, "User already exists with this username or email.", "", values)
			}
		} else {
			scheduler.renderAddUserFormGin(c, "Failed to add user: "+err.Error(), "", values)
		}
		return
	}

	// If administrator checkbox was checked, update the user to be an admin
	if administrator {
		// Get the newly created user to get their ID
		newUser, err := scheduler.GetUserByUsername(username)
		if err == nil && newUser != nil {
			err = scheduler.UpdateUserByID(newUser.ID, username, email, false, true)
			if err != nil {
				session := sessions.Default(c)
				session.Set("error", "User created but failed to set administrator privileges")
				session.Save()
				c.Redirect(http.StatusFound, "/scheduler/users")
				return
			}
		}
	}

	session := sessions.Default(c)
	session.Set("success", "User '"+username+"' added successfully")
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/users")
}

// AddRoomGin handles adding a new room
func (scheduler *wmu_scheduler) AddRoomGin(c *gin.Context) {
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	session := sessions.Default(c)

	// Get form data
	building := c.PostForm("building")
	roomNumber := c.PostForm("room_number")
	capacityStr := c.PostForm("capacity")
	computerLab := c.PostForm("computer_lab") == "on"
	dedicatedLab := c.PostForm("dedicated_lab") == "on"

	// Validate required fields
	if building == "" || roomNumber == "" || capacityStr == "" {
		session.Set("error", "Building, room number, and capacity are required")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_room")
		return
	}

	capacity, err := strconv.Atoi(capacityStr)
	if err != nil || capacity < 0 {
		session.Set("error", "Invalid capacity value")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_room")
		return
	}

	// Add the room
	err = scheduler.AddRoom(building, roomNumber, capacity, computerLab, dedicatedLab)
	if err != nil {
		session.Set("error", "Error adding room: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_room")
		return
	}

	session.Set("success", fmt.Sprintf("Room %s %s added successfully", building, roomNumber))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/rooms")
}

// SaveDepartmentsGin handles saving department changes and bulk deletion
func (scheduler *wmu_scheduler) SaveDepartmentsGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	session := sessions.Default(c)

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle department deletion
		departmentIDs := c.PostFormArray("department_ids[]")
		if len(departmentIDs) == 0 {
			session.Set("error", "No departments selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/departments")
			return
		}

		var errors []string
		deletedCount := 0

		for _, departmentIDStr := range departmentIDs {
			departmentID, err := strconv.Atoi(departmentIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid department ID: %s", departmentIDStr))
				continue
			}

			err = scheduler.DeleteDepartment(departmentID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete department ID %d: %v", departmentID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session.Set("error", fmt.Sprintf("%d departments deleted, %d errors occurred", deletedCount, len(errors)))
		} else {
			session.Set("success", fmt.Sprintf("%d departments deleted successfully", deletedCount))
		}
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	// Check for special cases
	if c.PostForm("no_changes") == "true" {
		session.Set("error", "No changes to save")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	if c.PostForm("no_selection") == "true" {
		session.Set("error", "No departments selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	// Get the departments JSON data
	departmentsJSON := c.PostForm("departments")
	if departmentsJSON == "" {
		session.Set("error", "No department data provided")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	// Parse the JSON data
	var departments []map[string]interface{}
	err = json.Unmarshal([]byte(departmentsJSON), &departments)
	if err != nil {
		session.Set("error", "Invalid department data format")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	// Update each department
	successCount := 0
	errorCount := 0

	for _, departmentData := range departments {
		id := getIntFromInterface(departmentData["id"])
		name := getStringFromInterface(departmentData["name"])

		if id <= 0 || name == "" {
			errorCount++
			continue
		}

		err = scheduler.UpdateDepartment(id, name)
		if err != nil {
			errorCount++
			continue
		}
		successCount++
	}

	// Set appropriate success/error message
	if errorCount > 0 {
		session.Set("error", fmt.Sprintf("%d departments updated, %d errors occurred", successCount, errorCount))
	} else {
		session.Set("success", fmt.Sprintf("%d departments updated successfully", successCount))
	}
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/departments")
}

// RenderAddDepartmentPageGin renders the add department page
func (scheduler *wmu_scheduler) RenderAddDepartmentPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	session.Delete("error")
	session.Save()

	data := gin.H{
		"User":      user,
		"CSRFToken": csrf.GetToken(c),
	}

	if errorMsg != nil {
		data["Error"] = errorMsg
	}

	c.HTML(http.StatusOK, "add_department", data)
}

// AddDepartmentGin handles adding a new department
func (scheduler *wmu_scheduler) AddDepartmentGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/departments")
		return
	}

	session := sessions.Default(c)

	// Get form data
	name := c.PostForm("name")

	// Validate required fields
	if name == "" {
		session.Set("error", "Department name is required")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_department")
		return
	}

	// Add the department
	err = scheduler.AddDepartment(name)
	if err != nil {
		session.Set("error", "Error adding department: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_department")
		return
	}

	session.Set("success", fmt.Sprintf("Department '%s' added successfully", name))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/departments")
}

// SavePrefixesGin handles saving prefix changes and bulk deletion
func (scheduler *wmu_scheduler) SavePrefixesGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	session := sessions.Default(c)

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle prefix deletion
		prefixIDs := c.PostFormArray("prefix_ids[]")
		if len(prefixIDs) == 0 {
			session.Set("error", "No prefixes selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/prefixes")
			return
		}

		var errors []string
		deletedCount := 0

		for _, prefixIDStr := range prefixIDs {
			prefixID, err := strconv.Atoi(prefixIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid prefix ID: %s", prefixIDStr))
				continue
			}

			err = scheduler.DeletePrefix(prefixID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete prefix ID %d: %v", prefixID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session.Set("error", fmt.Sprintf("%d prefixes deleted, %d errors occurred", deletedCount, len(errors)))
		} else {
			session.Set("success", fmt.Sprintf("%d prefixes deleted successfully", deletedCount))
		}
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	// Check for special cases
	if c.PostForm("no_changes") == "true" {
		session.Set("error", "No changes to save")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	if c.PostForm("no_selection") == "true" {
		session.Set("error", "No prefixes selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	// Get the prefixes JSON data
	prefixesJSON := c.PostForm("prefixes")
	if prefixesJSON == "" {
		session.Set("error", "No prefix data provided")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	// Parse the JSON data
	var prefixes []map[string]interface{}
	err = json.Unmarshal([]byte(prefixesJSON), &prefixes)
	if err != nil {
		session.Set("error", "Invalid prefix data format")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/prefixes")
		return
	}

	// Update each prefix
	successCount := 0
	errorCount := 0

	for _, prefixData := range prefixes {
		id := getIntFromInterface(prefixData["id"])
		prefix := getStringFromInterface(prefixData["prefix"])
		departmentID := getIntFromInterface(prefixData["department_id"])

		if id <= 0 || prefix == "" || departmentID <= 0 {
			errorCount++
			continue
		}

		// Get department name from ID
		departments, err := scheduler.GetAllDepartments()
		if err != nil {
			errorCount++
			continue
		}

		var departmentName string
		for _, dept := range departments {
			if dept.ID == departmentID {
				departmentName = dept.Name
				break
			}
		}

		if departmentName == "" {
			errorCount++
			continue
		}

		err = scheduler.UpdatePrefix(id, prefix, departmentName)
		if err != nil {
			errorCount++
			continue
		}
		successCount++
	}

	// Set appropriate success/error message
	if errorCount > 0 {
		session.Set("error", fmt.Sprintf("%d prefixes updated, %d errors occurred", successCount, errorCount))
	} else {
		session.Set("success", fmt.Sprintf("%d prefixes updated successfully", successCount))
	}
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/prefixes")
}

// RenderAddPrefixPageGin renders the add prefix page
func (scheduler *wmu_scheduler) RenderAddPrefixPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"Error": "Access denied. Administrator privileges required.",
			"User":  user,
		})
		return
	}

	// Get departments for the dropdown
	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	successMsg := session.Get("success")
	session.Delete("error")
	session.Delete("success")
	session.Save()

	data := gin.H{
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
		"Departments": departments,
		"Values":      gin.H{}, // Empty values for new prefix
	}

	if errorMsg != nil {
		data["Error"] = errorMsg
	}
	if successMsg != nil {
		data["Success"] = successMsg
	}

	c.HTML(http.StatusOK, "add_prefix", data)
}

// AddPrefixGin handles adding a new prefix
func (scheduler *wmu_scheduler) AddPrefixGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_prefix")
		return
	}

	session := sessions.Default(c)
	prefix := strings.TrimSpace(strings.ToUpper(c.PostForm("prefix")))
	departmentName := strings.TrimSpace(c.PostForm("department"))

	// Preserve form values for re-display on error
	values := gin.H{
		"prefix":     prefix,
		"department": departmentName,
	}

	// Validation
	if prefix == "" {
		session.Set("error", "Prefix code is required")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	if departmentName == "" {
		session.Set("error", "Department selection is required")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	// Validate prefix format (letters only, 2-10 characters)
	if !regexp.MustCompile(`^[A-Z]{2,10}$`).MatchString(prefix) {
		session.Set("error", "Prefix code must be 2-10 letters only (no numbers or special characters)")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	// Check if prefix already exists
	existingPrefixes, err := scheduler.GetAllPrefixes()
	if err != nil {
		AppLogger.LogError("Error checking existing prefixes", err)
		session.Set("error", "Error checking existing prefixes")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	for _, existingPrefix := range existingPrefixes {
		if existingPrefix.Prefix == prefix {
			session.Set("error", fmt.Sprintf("Prefix '%s' already exists", prefix))
			session.Save()
			scheduler.renderAddPrefixFormWithValues(c, values)
			return
		}
	}

	// Verify department exists
	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		AppLogger.LogError("Error loading departments for validation", err)
		session.Set("error", "Error validating department")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	var departmentExists bool
	for _, dept := range departments {
		if dept.Name == departmentName {
			departmentExists = true
			break
		}
	}

	if !departmentExists {
		session.Set("error", "Selected department does not exist")
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	// Add the prefix
	err = scheduler.AddPrefix(prefix, departmentName)
	if err != nil {
		AppLogger.LogError("Error adding prefix", err)
		session.Set("error", fmt.Sprintf("Error adding prefix: %v", err))
		session.Save()
		scheduler.renderAddPrefixFormWithValues(c, values)
		return
	}

	// Success
	AppLogger.LogInfo(fmt.Sprintf("User %s added new prefix: %s (%s)", user.Username, prefix, departmentName))
	session.Set("success", fmt.Sprintf("Prefix '%s' added successfully", prefix))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/prefixes")
}

// renderAddPrefixFormWithValues renders the add prefix form with preserved values
func (scheduler *wmu_scheduler) renderAddPrefixFormWithValues(c *gin.Context, values gin.H) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get departments for the dropdown
	departments, err := scheduler.GetAllDepartments()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error loading departments: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	errorMsg := session.Get("error")
	successMsg := session.Get("success")
	session.Delete("error")
	session.Delete("success")
	session.Save()

	data := gin.H{
		"User":        user,
		"CSRFToken":   csrf.GetToken(c),
		"Departments": departments,
		"Values":      values,
	}

	if errorMsg != nil {
		data["Error"] = errorMsg
	}
	if successMsg != nil {
		data["Success"] = successMsg
	}

	c.HTML(http.StatusOK, "add_prefix", data)
}

// SaveUsersGin handles saving user changes and bulk deletion
func (scheduler *wmu_scheduler) SaveUsersGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is administrator
	if !user.Administrator {
		session := sessions.Default(c)
		session.Set("error", "Access denied. Administrator privileges required.")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	session := sessions.Default(c)

	// Check if this is a delete operation
	action := c.PostForm("action")
	if action == "delete" {
		// Handle user deletion
		userIDs := c.PostFormArray("user_ids[]")
		if len(userIDs) == 0 {
			session.Set("error", "No users selected for deletion")
			session.Save()
			c.Redirect(http.StatusFound, "/scheduler/users")
			return
		}

		var errors []string
		deletedCount := 0

		for _, userIDStr := range userIDs {
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid user ID: %s", userIDStr))
				continue
			}

			// Don't allow deleting the current user
			if userID == user.ID {
				errors = append(errors, "Cannot delete your own account")
				continue
			}

			err = scheduler.DeleteUserByID(userID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to delete user ID %d: %v", userID, err))
				continue
			}
			deletedCount++
		}

		if len(errors) > 0 {
			session.Set("error", fmt.Sprintf("%d users deleted, %d errors occurred", deletedCount, len(errors)))
		} else {
			session.Set("success", fmt.Sprintf("%d users deleted successfully", deletedCount))
		}
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	// Check for special cases
	if c.PostForm("no_changes") == "true" {
		session.Set("error", "No changes to save")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	if c.PostForm("no_selection") == "true" {
		session.Set("error", "No users selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	// Get the users JSON data
	usersJSON := c.PostForm("users")
	if usersJSON == "" {
		session.Set("error", "No user data provided")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	// Parse the JSON data
	var users []map[string]interface{}
	err = json.Unmarshal([]byte(usersJSON), &users)
	if err != nil {
		session.Set("error", "Invalid user data format")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/users")
		return
	}

	// Update each user
	successCount := 0
	errorCount := 0

	for _, userData := range users {
		id := getIntFromInterface(userData["id"])
		username := getStringFromInterface(userData["username"])
		email := getStringFromInterface(userData["email"])
		administrator := userData["administrator"] == true || userData["administrator"] == "true"

		if id <= 0 || username == "" || email == "" {
			errorCount++
			continue
		}

		err = scheduler.UpdateUserByID(id, username, email, false, administrator)
		if err != nil {
			errorCount++
			continue
		}
		successCount++
	}

	// Set appropriate success/error message
	if errorCount > 0 {
		session.Set("error", fmt.Sprintf("%d users updated, %d errors occurred", successCount, errorCount))
	} else {
		session.Set("success", fmt.Sprintf("%d users updated successfully", successCount))
	}
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/users")
}

// SetErrorMessageGin handles setting error messages in session
func (scheduler *wmu_scheduler) SetErrorMessageGin(c *gin.Context) {
	message := c.PostForm("message")
	if message != "" {
		session := sessions.Default(c)
		session.Set("error", message)
		session.Save()
	}
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ExportCoursesToExcel exports all courses for a given schedule to Excel format
func (scheduler *wmu_scheduler) ExportCoursesToExcel(c *gin.Context) {
	// Get schedule ID from URL parameter
	scheduleID := c.PostForm("schedule_id")
	if scheduleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule ID is required"})
		return
	}

	scheduleIDInt, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	// Get schedule information
	schedule, err := scheduler.GetScheduleByID(scheduleIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve schedule"})
		return
	}

	// Get courses for the schedule
	courses, err := scheduler.GetActiveCoursesForSchedule(scheduleIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve courses"})
		return
	}

	// Get lookup data for references
	instructors, err := scheduler.GetAllInstructors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve instructors"})
		return
	}
	instructorMap := make(map[int]Instructor)
	for _, instructor := range instructors {
		instructorMap[instructor.ID] = instructor
	}

	timeslots, err := scheduler.GetAllTimeSlots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timeslots"})
		return
	}
	timeslotMap := make(map[int]TimeSlot)
	for _, timeslot := range timeslots {
		timeslotMap[timeslot.ID] = timeslot
	}

	rooms, err := scheduler.GetAllRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rooms"})
		return
	}
	roomMap := make(map[int]Room)
	for _, room := range rooms {
		roomMap[room.ID] = room
	}

	// Create new Excel file
	f := excelize.NewFile()
	sheetName := schedule.Term + " " + fmt.Sprintf("%d", schedule.Year)

	// Set the default sheet name
	f.SetSheetName("Sheet1", sheetName)

	// Row 1: Merged header with schedule info (A1:E1)
	headerText := fmt.Sprintf("%s %s %d", schedule.Prefix, schedule.Term, schedule.Year)
	f.SetCellValue(sheetName, "A1", headerText)
	f.MergeCell(sheetName, "A1", "E1")

	// Style for row 1 - brown text, white background, 14pt bold
	row1Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  14,
			Color: "8B4513", // Brown color
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFFFFF"}, // White background
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A1", "E1", row1Style)

	// Row 2: Column labels
	f.SetCellValue(sheetName, "A2", "Term")
	f.SetCellValue(sheetName, "B2", "College")
	f.SetCellValue(sheetName, "C2", "Department")
	f.SetCellValue(sheetName, "D2", "Subject")
	f.SetCellValue(sheetName, "E2", "Campus")

	// Style for row 2 - tan background with bold black text
	row2Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "000000", // Black color
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"D2B48C"}, // Tan background
			Pattern: 1,
		},
	})
	f.SetCellStyle(sheetName, "A2", "E2", row2Style)

	// Row 3: Values
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("%s %d", schedule.Term, schedule.Year))
	f.SetCellValue(sheetName, "B3", "Engineering & Applied Sciences")
	f.SetCellValue(sheetName, "C3", schedule.Department)
	f.SetCellValue(sheetName, "D3", schedule.Prefix)
	f.SetCellValue(sheetName, "E3", "Main")

	// Style for row 3 - bold black text
	row3Style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "000000", // Black color
		},
	})
	f.SetCellStyle(sheetName, "A3", "E3", row3Style)

	// Define headers based on the import template structure (with removed columns)
	headers := []string{
		"CRN", "Course ID", "Section", "Title", "Lab", "Credit Hours", "Contact Hours",
		"Cap", "Spec Appr", "Mtg Type", "Days", "Time", "Location",
		"Primary Instructor", "Comment ",
	}

	// Write headers to row 5 (Excel row numbering starts at 1)
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, 5)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style the header row
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "000000", // Black color
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFFDD0"}, // Light cream background
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Apply header style to all header columns first
	f.SetCellStyle(sheetName, "A5", "O5", headerStyle)

	// Create center alignment style for data rows only
	centerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// Create center alignment style with header formatting for header row
	centerHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "000000", // Black color
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFFDD0"}, // Light cream background
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// Apply center alignment to specified columns
	centerColumns := []string{"A", "B", "C", "E", "F", "G", "H", "I", "J", "K", "L", "M"} // CRN, Course ID, Section, Lab, Credit Hours, Contact Hours, Cap, Spec Appr, Mtg Type, Days, Time, Location
	for _, col := range centerColumns {
		// Apply center header style to header row
		f.SetCellStyle(sheetName, fmt.Sprintf("%s5", col), fmt.Sprintf("%s5", col), centerHeaderStyle)
		// Apply center style to data rows
		if len(courses) > 0 {
			f.SetCellStyle(sheetName, fmt.Sprintf("%s6", col), fmt.Sprintf("%s%d", col, 6+len(courses)-1), centerStyle)
		}
	}

	// Write course data starting from row 6
	for i, course := range courses {
		row := i + 6 // Start from row 6

		// Helper function to format time from TimeSlot
		formatTime := func(timeslot TimeSlot) string {
			if timeslot.StartTime == "" || timeslot.EndTime == "" {
				return ""
			}
			return fmt.Sprintf("%s - %s", timeslot.StartTime, timeslot.EndTime)
		}

		// Helper function to format days from TimeSlot
		formatDays := func(timeslot TimeSlot) string {
			var days []string
			if timeslot.Monday {
				days = append(days, "M")
			}
			if timeslot.Tuesday {
				days = append(days, "T")
			}
			if timeslot.Wednesday {
				days = append(days, "W")
			}
			if timeslot.Thursday {
				days = append(days, "R")
			}
			if timeslot.Friday {
				days = append(days, "F")
			}
			return strings.Join(days, "")
		}

		// Helper function to format instructor name
		formatInstructor := func(instructorID int) string {
			if instructor, exists := instructorMap[instructorID]; exists {
				return fmt.Sprintf("%s, %s", instructor.LastName, instructor.FirstName)
			}
			return ""
		}

		// Helper function to format room location
		formatLocation := func(roomID int) string {
			if room, exists := roomMap[roomID]; exists {
				return fmt.Sprintf("%s %s", room.Building, room.RoomNumber)
			}
			return ""
		}

		// Get related data
		var timeslot TimeSlot

		if course.TimeSlotID != -1 {
			timeslot = timeslotMap[course.TimeSlotID]
		}

		// Build course ID (Prefix + Course Number)
		courseID := fmt.Sprintf("%s %s", course.Prefix, course.CourseNumber)

		// Helper function to format credit hours range
		formatCredits := func(min, max int) string {
			if max > min {
				return fmt.Sprintf("%d-%d", min, max)
			}
			return fmt.Sprintf("%d", min)
		}

		// Helper function to format contact hours range
		formatContact := func(min, max int) string {
			if max > min {
				return fmt.Sprintf("%d-%d", min, max)
			}
			return fmt.Sprintf("%d", min)
		}

		// Set cell values (adjusted for new column structure)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), course.CRN)     // CRN
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), courseID)       // Course ID
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), course.Section) // Section
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), course.Title)   // Title
		if course.Lab {                                                    // Lab
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "✓")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), "")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), formatCredits(course.MinCredits, course.MaxCredits)) // Credit Hours
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), formatContact(course.MinContact, course.MaxContact)) // Contact Hours
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), course.Cap)                                          // Cap
		if course.Approval {                                                                                    // Spec Appr
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "✓")
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), "")
		}
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), course.Mode)                           // Mtg Type
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), formatDays(timeslot))                  // Days
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), formatTime(timeslot))                  // Time
		f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), formatLocation(course.RoomID))         // Location
		f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), formatInstructor(course.InstructorID)) // Primary Instructor
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), course.Comment)                        // Comment
	}

	// Set custom column widths
	f.SetColWidth(sheetName, "A", "A", 15) // CRN
	f.SetColWidth(sheetName, "B", "B", 30) // Course ID
	f.SetColWidth(sheetName, "C", "C", 20) // Section
	f.SetColWidth(sheetName, "D", "D", 30) // Title (increased to 30)
	f.SetColWidth(sheetName, "E", "E", 12) // Lab
	f.SetColWidth(sheetName, "F", "F", 10) // Credit Hours
	f.SetColWidth(sheetName, "G", "G", 12) // Contact Hours
	f.SetColWidth(sheetName, "H", "H", 8)  // Cap
	f.SetColWidth(sheetName, "I", "I", 10) // Spec Appr
	f.SetColWidth(sheetName, "J", "J", 12) // Mtg Type
	f.SetColWidth(sheetName, "K", "K", 8)  // Days
	f.SetColWidth(sheetName, "L", "L", 30) // Time (increased to 30)
	f.SetColWidth(sheetName, "M", "M", 15) // Location
	f.SetColWidth(sheetName, "N", "N", 20) // Primary Instructor
	f.SetColWidth(sheetName, "O", "O", 15) // Comment

	// Generate filename with schedule info
	filename := fmt.Sprintf("%s_%s_%d.xlsx", schedule.Prefix, schedule.Term, schedule.Year)

	// Set response headers for file download
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// Write file to response
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
		return
	}
}

// ScheduleData represents the organized schedule data for the template
type ScheduleData struct {
	Monday    map[string][]CourseScheduleItem
	Tuesday   map[string][]CourseScheduleItem
	Wednesday map[string][]CourseScheduleItem
	Thursday  map[string][]CourseScheduleItem
	Friday    map[string][]CourseScheduleItem
}

func timeStringToMinutes(timeStr string) int {
	// Parse the time string (e.g., "8:30 AM") into a time.Time object
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return -1
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return -1
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1
	}
	return hour*60 + minute
}

func addCourseInRange(dayMap map[string][]CourseScheduleItem, course CourseScheduleItem) {
	var startTime, endTime int
	// Convert time strings to integers for easier comparison
	if course.StartTime == "" || course.EndTime == "" {
		return // Skip if times are not set
	}
	// Convert course.StartTime (e.g., "8:30 AM") to minutes since midnight
	startTime = timeStringToMinutes(course.StartTime)
	endTime = timeStringToMinutes(course.EndTime)

	for t := startTime; t < endTime; t += 30 {
		hour := t / 60
		minute := t % 60
		ampm := "AM"
		displayHour := hour
		if hour == 0 {
			displayHour = 12
		} else if hour > 12 {
			displayHour = hour - 12
			ampm = "PM"
		} else if hour == 12 {
			ampm = "PM"
		}
		timeStr := fmt.Sprintf("%d:%02d %s", displayHour, minute, ampm)
		dayMap[timeStr] = append(dayMap[timeStr], course)
	}
}

// RenderCoursesTableGin renders the courses table page
func (scheduler *wmu_scheduler) RenderCoursesTableGin(c *gin.Context) {
	session := sessions.Default(c)

	// Get all courses with time slot and instructor data
	courseScheduleItems, err := scheduler.GetCoursesWithScheduleData()
	if err != nil {
		AppLogger.Printf("Error getting courses with schedule data: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Unable to load course schedule data",
		})
		return
	}
	timeSlotStrings := make([]string, 0)
	timeSlotStrings = append(timeSlotStrings, "8:00 AM", "8:30 AM", "9:00 AM", "9:30 AM", "10:00 AM", "10:30 AM",
		"11:00 AM", "11:30 AM", "12:00 PM", "12:30 PM", "1:00 PM", "1:30 PM", "2:00 PM", "2:30 PM", "3:00 PM",
		"3:30 PM", "4:00 PM", "4:30 PM", "5:00 PM", "5:30 PM", "6:00 PM", "6:30 PM", "7:00 PM", "7:30 PM",
		"8:00 PM", "8:30 PM", "9:00 PM", "9:30 PM")

	// Initialize schedule data structure with all time slots
	schedule := ScheduleData{
		Monday:    make(map[string][]CourseScheduleItem),
		Tuesday:   make(map[string][]CourseScheduleItem),
		Wednesday: make(map[string][]CourseScheduleItem),
		Thursday:  make(map[string][]CourseScheduleItem),
		Friday:    make(map[string][]CourseScheduleItem),
	}

	// Pre-populate all time slots with empty slices
	for _, timeSlot := range timeSlotStrings {
		schedule.Monday[timeSlot] = []CourseScheduleItem{}
		schedule.Tuesday[timeSlot] = []CourseScheduleItem{}
		schedule.Wednesday[timeSlot] = []CourseScheduleItem{}
		schedule.Thursday[timeSlot] = []CourseScheduleItem{}
		schedule.Friday[timeSlot] = []CourseScheduleItem{}
	}

	// Organize courses by day and time
	for _, course := range courseScheduleItems {
		if course.Monday {
			addCourseInRange(schedule.Monday, course)
		}
		if course.Tuesday {
			addCourseInRange(schedule.Tuesday, course)
		}
		if course.Wednesday {
			addCourseInRange(schedule.Wednesday, course)
		}
		if course.Thursday {
			addCourseInRange(schedule.Thursday, course)
		}
		if course.Friday {
			addCourseInRange(schedule.Friday, course)
		}
	}

	// Get any session messages
	var errorMsg, successMsg string
	if msg := session.Get("error"); msg != nil {
		errorMsg = msg.(string)
		session.Delete("error")
	}
	if msg := session.Get("success"); msg != nil {
		successMsg = msg.(string)
		session.Delete("success")
	}
	session.Save()

	c.HTML(http.StatusOK, "courses_table", gin.H{
		"TimeSlots": timeSlotStrings,
		"Schedule":  schedule,
		"Error":     errorMsg,
		"Success":   successMsg,
		"CSRFToken": csrf.GetToken(c),
	})
}

// Conflict detection structures
type ConflictPair struct {
	Course1 CourseDetail
	Course2 CourseDetail
	Type    string // "instructor" or "room"
}

type CourseDetail struct {
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
	TimeSlot     *TimeSlot
}

type ConflictReport struct {
	InstructorConflicts []ConflictPair
	RoomConflicts       []ConflictPair
	Schedule1ID         int
	Schedule2ID         int
}

// DetectScheduleConflictsGin detects conflicts between two schedules
func (scheduler *wmu_scheduler) DetectScheduleConflictsGin(c *gin.Context) {
	// Get current user for authorization
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	schedule1ID := c.PostForm("schedule1_id")
	schedule2ID := c.PostForm("schedule2_id")

	if schedule1ID == "" || schedule2ID == "" {
		session := sessions.Default(c)
		session.Set("error", "Both schedules must be selected")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/conflicts")
		return
	}

	id1, err := strconv.Atoi(schedule1ID)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid schedule selection")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/conflicts")
		return
	}

	id2, err := strconv.Atoi(schedule2ID)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid schedule selection")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/conflicts")
		return
	}

	conflicts, err := scheduler.DetectConflictsBetweenSchedules(id1, id2)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Failed to detect conflicts: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/conflicts")
		return
	}

	// Get schedule names for display
	schedules, err := scheduler.GetAllSchedules()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching schedules: " + err.Error(),
			"User":  user,
		})
		return
	}

	var schedule1Name, schedule2Name string
	for _, sched := range schedules {
		if sched.ID == id1 {
			schedule1Name = fmt.Sprintf("%s %s %d", sched.Prefix, sched.Term, sched.Year)
		}
		if sched.ID == id2 {
			schedule2Name = fmt.Sprintf("%s %s %d", sched.Prefix, sched.Term, sched.Year)
		}
	}

	c.HTML(http.StatusOK, "conflict_display.html", gin.H{
		"User":          user,
		"Conflicts":     conflicts,
		"Schedule1Name": schedule1Name,
		"Schedule2Name": schedule2Name,
		"CSRFToken":     csrf.GetToken(c),
	})
}

// RenderConflictSelectPageGin renders the conflict selection page
func (scheduler *wmu_scheduler) RenderConflictSelectPageGin(c *gin.Context) {
	// Get current user for authorization
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get all schedules for dropdown
	schedules, err := scheduler.GetAllSchedules()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching schedules: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Get any session messages
	session := sessions.Default(c)
	var errorMsg, successMsg string
	if msg := session.Get("error"); msg != nil {
		errorMsg = msg.(string)
		session.Delete("error")
	}
	if msg := session.Get("success"); msg != nil {
		successMsg = msg.(string)
		session.Delete("success")
	}
	session.Save()

	c.HTML(http.StatusOK, "conflict_select.html", gin.H{
		"User":      user,
		"Schedules": schedules,
		"Error":     errorMsg,
		"Success":   successMsg,
		"CSRFToken": csrf.GetToken(c),
	})
}
func conflictExists(conflicts []ConflictPair, c1, c2 CourseDetail, conflictType string) bool {
	for _, pair := range conflicts {
		if pair.Type != conflictType {
			continue
		}
		// Check both (c1, c2) and (c2, c1)
		if (pair.Course1.ID == c1.ID && pair.Course2.ID == c2.ID) ||
			(pair.Course1.ID == c2.ID && pair.Course2.ID == c1.ID) {
			return true
		}
	}
	return false
}

// DetectConflictsBetweenSchedules performs the actual conflict detection logic
func (scheduler *wmu_scheduler) DetectConflictsBetweenSchedules(schedule1ID, schedule2ID int) (*ConflictReport, error) {
	// Get courses from both schedules with detailed information
	courses1, err := scheduler.getCoursesWithDetails(schedule1ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses for schedule %d: %v", schedule1ID, err)
	}

	courses2, err := scheduler.getCoursesWithDetails(schedule2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses for schedule %d: %v", schedule2ID, err)
	}

	var instructorConflicts []ConflictPair
	var roomConflicts []ConflictPair

	// Compare each course from schedule1 with each course from schedule2
	for _, course1 := range courses1 {
		for _, course2 := range courses2 {
			// Skip identical courses if comparing the same schedule
			if schedule1ID == schedule2ID && course1.ID == course2.ID {
				continue
			}

			crosslist, err := scheduler.AreCoursesCrosslisted(course1.CRN, course2.CRN)
			if err != nil || crosslist {
				// Skip crosslisted courses
				continue
			}

			// Check if time slots overlap
			if scheduler.timeSlotsOverlap(course1.TimeSlot, course2.TimeSlot) {
				// Check for instructor conflicts
				if course1.InstructorID == course2.InstructorID && course1.InstructorID > 0 {
					// Check for FSO/PSO exception
					if !scheduler.isFSOPSOException(course1, course2) {
						conflictPair := ConflictPair{
							Course1: course1,
							Course2: course2,
							Type:    "instructor",
						}
						// Avoid duplicate conflicts
						if !conflictExists(instructorConflicts, course1, course2, "instructor") {

							// Add to instructor conflicts
							instructorConflicts = append(instructorConflicts, conflictPair)
						}
					}
				}

				// Check for room conflicts (different courses in same room)
				if course1.RoomID == course2.RoomID && course1.RoomID > 0 && !scheduler.isSameCourse(course1, course2) {
					conflictPair := ConflictPair{
						Course1: course1,
						Course2: course2,
						Type:    "room",
					}
					// Avoid duplicate conflicts
					if !conflictExists(roomConflicts, course1, course2, "room") {
						// Add to room conflicts
						roomConflicts = append(roomConflicts, conflictPair)
					}
				}
			}
		}
	}

	return &ConflictReport{
		InstructorConflicts: instructorConflicts,
		RoomConflicts:       roomConflicts,
		Schedule1ID:         schedule1ID,
		Schedule2ID:         schedule2ID,
	}, nil
}

// getCoursesWithDetails retrieves courses with their timeslot details
func (scheduler *wmu_scheduler) getCoursesWithDetails(scheduleID int) ([]CourseDetail, error) {
	courses, err := scheduler.GetActiveCoursesForSchedule(scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active courses for schedule %d: %v", scheduleID, err)
	}

	courseDetail := make([]CourseDetail, 0)

	for _, course := range courses {

		timeslot, err := scheduler.GetTimeSlotById(course.TimeSlotID)
		if err != nil {
			return nil, fmt.Errorf("failed to get timeslot for course %d: %v", course.ID, err)
		}

		// Populate the TimeSlot information
		courseDetail = append(courseDetail, CourseDetail{
			ID:           course.ID,
			CRN:          course.CRN,
			Section:      course.Section,
			ScheduleID:   course.ScheduleID,
			Prefix:       course.Prefix,
			CourseNumber: course.CourseNumber,
			Title:        course.Title,
			InstructorID: course.InstructorID,
			TimeSlotID:   course.TimeSlotID,
			RoomID:       course.RoomID,
			Mode:         course.Mode,
			TimeSlot:     timeslot,
		})

	}

	return courseDetail, nil
}

// timeSlotsOverlap checks if two time slots overlap in both time and days
func (scheduler *wmu_scheduler) timeSlotsOverlap(ts1, ts2 *TimeSlot) bool {
	if ts1 == nil || ts2 == nil {
		return false // If either timeslot is nil, no overlap
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

	// Check if times overlap
	// Convert time strings to comparable format (assuming HH:MM format)
	return scheduler.timeRangesOverlap(ts1.StartTime, ts1.EndTime, ts2.StartTime, ts2.EndTime)
}

// timeRangesOverlap checks if two time ranges overlap
func (scheduler *wmu_scheduler) timeRangesOverlap(start1, end1, start2, end2 string) bool {
	// If any time is empty, assume no overlap
	if start1 == "" || end1 == "" || start2 == "" || end2 == "" {
		return false
	}

	// Simple string comparison should work for HH:MM format
	// start1 < end2 && start2 < end1
	return start1 < end2 && start2 < end1
}

// isFSOPSOException checks if courses are exempt from instructor conflicts due to FSO/PSO mode
func (scheduler *wmu_scheduler) isFSOPSOException(course1, course2 CourseDetail) bool {
	// If courses have the same course number and one is FSO or PSO, no conflict
	if course1.Prefix == course2.Prefix && course1.CourseNumber == course2.CourseNumber {
		return (course1.Mode == "FSO" || course1.Mode == "PSO") ||
			(course2.Mode == "FSO" || course2.Mode == "PSO")
	}
	return false
}

// isSameCourse checks if two courses are the same course (same prefix and course number)
func (scheduler *wmu_scheduler) isSameCourse(course1, course2 CourseDetail) bool {
	return course1.Prefix == course2.Prefix && course1.CourseNumber == course2.CourseNumber
}

// CrosslistingDisplayItem represents a cross-listing with enriched course and schedule data for display
type CrosslistingDisplayItem struct {
	ID        int
	Course1   CourseDetail
	Course2   CourseDetail
	Schedule1 Schedule
	Schedule2 Schedule
	CreatedAt string
	UpdatedAt string
}

// RenderCrosslistingsPageGin renders the crosslistings page for the current schedule
func (scheduler *wmu_scheduler) RenderCrosslistingsPageGin(c *gin.Context) {
	// Check authentication
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Get schedule_id from session
	scheduleIDStr, err := scheduler.getCurrentSchedule(c)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "No schedule selected. Please select a schedule first.",
		})
		return
	}

	// Convert schedule ID to integer
	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid schedule ID",
		})
		return
	}

	// Get current schedule information
	currentSchedule, err := scheduler.GetScheduleByID(scheduleID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to get schedule information: " + err.Error(),
		})
		return
	}

	if currentSchedule == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"Error": "Schedule not found",
		})
		return
	}

	// Get all crosslistings for this schedule
	crosslistings, err := scheduler.GetAllCrosslistingsForSchedule(scheduleID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to get crosslistings: " + err.Error(),
		})
		return
	}

	// Enrich crosslistings with course and schedule details
	var enrichedCrosslistings []CrosslistingDisplayItem
	for _, cl := range crosslistings {
		var enriched CrosslistingDisplayItem
		enriched.ID = cl.ID
		enriched.CreatedAt = cl.CreatedAt
		enriched.UpdatedAt = cl.UpdatedAt

		// Get course details for CRN1
		course1, err := scheduler.getCourseDetailsByCRN(cl.CRN1)
		if err != nil {
			AppLogger.LogError(fmt.Sprintf("Failed to get course details for CRN %d", cl.CRN1), err)
			continue
		}
		enriched.Course1 = course1

		// Get course details for CRN2
		course2, err := scheduler.getCourseDetailsByCRN(cl.CRN2)
		if err != nil {
			AppLogger.LogError(fmt.Sprintf("Failed to get course details for CRN %d", cl.CRN2), err)
			continue
		}
		enriched.Course2 = course2

		// Get schedule details for Schedule1
		schedule1, err := scheduler.GetScheduleByID(cl.ScheduleID1)
		if err != nil || schedule1 == nil {
			AppLogger.LogError(fmt.Sprintf("Failed to get schedule details for ID %d", cl.ScheduleID1), err)
			continue
		}
		enriched.Schedule1 = *schedule1

		// Get schedule details for Schedule2
		schedule2, err := scheduler.GetScheduleByID(cl.ScheduleID2)
		if err != nil || schedule2 == nil {
			AppLogger.LogError(fmt.Sprintf("Failed to get schedule details for ID %d", cl.ScheduleID2), err)
			continue
		}
		enriched.Schedule2 = *schedule2

		enrichedCrosslistings = append(enrichedCrosslistings, enriched)
	}

	// Render the template
	c.HTML(http.StatusOK, "crosslistings.html", gin.H{
		"User":          user,
		"Schedule":      currentSchedule,
		"Crosslistings": enrichedCrosslistings,
		"Success":       successMsg,
		"Error":         errorMsg,
		"CSRFToken":     csrf.GetToken(c),
	})
}

// getCourseDetailsByCRN retrieves detailed course information by CRN
func (scheduler *wmu_scheduler) getCourseDetailsByCRN(crn int) (CourseDetail, error) {
	var course CourseDetail

	query := `
		SELECT c.id, c.crn, c.section, c.schedule_id, p.prefix, c.course_number, c.title,
			   COALESCE(c.instructor_id, -1) as instructor_id,
			   COALESCE(c.timeslot_id, -1) as timeslot_id,
			   COALESCE(c.room_id, -1) as room_id,
			   c.mode
		FROM courses c
		JOIN schedules s ON c.schedule_id = s.id
		JOIN prefixes p ON s.prefix_id = p.id
		WHERE c.crn = ? AND c.status != 'Deleted'
	`

	err := scheduler.database.QueryRow(query, crn).Scan(
		&course.ID, &course.CRN, &course.Section, &course.ScheduleID,
		&course.Prefix, &course.CourseNumber, &course.Title,
		&course.InstructorID, &course.TimeSlotID, &course.RoomID,
		&course.Mode,
	)

	if err != nil {
		return course, err
	}

	// Get time slot details if available
	if course.TimeSlotID != -1 {
		timeSlot, err := scheduler.GetTimeSlotById(course.TimeSlotID)
		if err == nil && timeSlot != nil {
			course.TimeSlot = timeSlot
		}
	}

	return course, nil
}

// RenderAddCrosslistingPageGin renders the add crosslisting form page
func (scheduler *wmu_scheduler) RenderAddCrosslistingPageGin(c *gin.Context) {
	// Check authentication
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get any error or success messages from session
	session := sessions.Default(c)
	successMsg := session.Get("success")
	errorMsg := session.Get("error")
	session.Delete("success")
	session.Delete("error")
	session.Save()

	// Get all schedules for dropdowns
	schedules, err := scheduler.GetAllSchedules()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to get schedules: " + err.Error(),
			"User":  user,
		})
		return
	}

	// Render the form template
	c.HTML(http.StatusOK, "add_crosslisting.html", gin.H{
		"User":      user,
		"Schedules": schedules,
		"Success":   successMsg,
		"Error":     errorMsg,
		"CSRFToken": csrf.GetToken(c),
	})
}

// AddCrosslistingGin handles the POST request to add a new crosslisting
func (scheduler *wmu_scheduler) AddCrosslistingGin(c *gin.Context) {
	// Check authentication
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get form data
	schedule1IDStr := c.PostForm("schedule1")
	course1IDStr := c.PostForm("course1")
	schedule2IDStr := c.PostForm("schedule2")
	course2IDStr := c.PostForm("course2")

	// Validate form data
	if schedule1IDStr == "" || course1IDStr == "" || schedule2IDStr == "" || course2IDStr == "" {
		session := sessions.Default(c)
		session.Set("error", "All fields are required")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	// Convert to integers
	schedule1ID, err := strconv.Atoi(schedule1IDStr)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid schedule 1 selection")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	schedule2ID, err := strconv.Atoi(schedule2IDStr)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid schedule 2 selection")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	// Parse CRNs from course selections (format: "crn:title")
	course1Parts := strings.SplitN(course1IDStr, ":", 2)
	course2Parts := strings.SplitN(course2IDStr, ":", 2)

	if len(course1Parts) < 1 || len(course2Parts) < 1 {
		session := sessions.Default(c)
		session.Set("error", "Invalid course selections")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	crn1, err := strconv.Atoi(course1Parts[0])
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid CRN for course 1")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	crn2, err := strconv.Atoi(course2Parts[0])
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid CRN for course 2")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	// Validate that the courses are different
	if crn1 == crn2 {
		session := sessions.Default(c)
		session.Set("error", "Cannot crosslist a course with itself")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	// Add the crosslisting to the database
	err = scheduler.AddOrUpdateCrosslisting(crn1, crn2, schedule1ID, schedule2ID)
	if err != nil {
		AppLogger.LogError("Failed to add crosslisting", err)
		session := sessions.Default(c)
		session.Set("error", "Failed to add crosslisting: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/add_crosslisting")
		return
	}

	// Success
	session := sessions.Default(c)
	session.Set("success", fmt.Sprintf("Successfully added crosslisting between CRN %d and CRN %d", crn1, crn2))
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/crosslistings")
}

// GetCoursesForScheduleAPIGin provides an API endpoint to get courses for a schedule (for AJAX calls)
func (scheduler *wmu_scheduler) GetCoursesForScheduleAPIGin(c *gin.Context) {
	// Check authentication
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Get schedule ID from URL parameter
	scheduleIDStr := c.Param("scheduleId")
	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid schedule ID"})
		return
	}

	// Get active courses for the schedule
	courses, err := scheduler.GetActiveCoursesForSchedule(scheduleID)
	if err != nil {
		AppLogger.LogError(fmt.Sprintf("Failed to get courses for schedule %d", scheduleID), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get courses"})
		return
	}

	// Format courses for the dropdown (CRN:Title format expected by frontend)
	type CourseOption struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	}

	var courseOptions []CourseOption
	for _, course := range courses {
		value := fmt.Sprintf("%d", course.CRN)
		text := fmt.Sprintf("CRN %d: %s %s - %s", course.CRN, course.Prefix, course.CourseNumber, course.Title)
		courseOptions = append(courseOptions, CourseOption{
			Value: value,
			Text:  text,
		})
	}

	c.JSON(http.StatusOK, courseOptions)
}

type CourseForCrosslist struct {
	CRN          int    `json:"crn"`
	Prefix       string `json:"prefix"`
	CourseNumber string `json:"course_number"`
	Days         string `json:"days"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

// DeleteCrosslistingGin handles deletion of a crosslisting
func (scheduler *wmu_scheduler) DeleteCrosslistingGin(c *gin.Context) {
	// Check authentication
	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get crosslisting ID from form
	crosslistingIDStr := c.PostForm("crosslisting_id")
	if crosslistingIDStr == "" {
		session := sessions.Default(c)
		session.Set("error", "No crosslisting selected for deletion")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/crosslistings")
		return
	}

	crosslistingID, err := strconv.Atoi(crosslistingIDStr)
	if err != nil {
		session := sessions.Default(c)
		session.Set("error", "Invalid crosslisting ID")
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/crosslistings")
		return
	}

	// Delete the crosslisting
	err = scheduler.DeleteCrosslisting(crosslistingID)
	if err != nil {
		AppLogger.LogError(fmt.Sprintf("Failed to delete crosslisting %d", crosslistingID), err)
		session := sessions.Default(c)
		session.Set("error", "Failed to delete crosslisting: "+err.Error())
		session.Save()
		c.Redirect(http.StatusFound, "/scheduler/crosslistings")
		return
	}

	// Success
	session := sessions.Default(c)
	session.Set("success", "Crosslisting deleted successfully")
	session.Save()
	c.Redirect(http.StatusFound, "/scheduler/crosslistings")
}
