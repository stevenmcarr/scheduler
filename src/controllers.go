package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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
	scheduler.renderSignupFormGin(c, "", "", nil)
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
	// Check for success message from URL parameters
	successMsg := c.Query("success")
	scheduler.renderLoginFormGin(c, "", successMsg, nil)
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

	// Preserve form values for re-display
	values := map[string]string{
		"username": username,
		"email":    email,
	}

	if username == "" || email == "" || password == "" {
		scheduler.renderSignupFormGin(c, "All fields are required", "", values)
		return
	}

	// Use email as username since that's what the database expects
	err := scheduler.AddUser(username, email, password)
	if err != nil {
		scheduler.renderSignupFormGin(c, err.Error(), "", values)
		return
	}

	// Show success message and redirect
	c.Redirect(http.StatusFound, "/scheduler/login?success=Account created successfully")
}

func (scheduler *wmu_scheduler) RenderCoursesPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get schedule_id from the URL query parameters
	scheduleID := c.Query("schedule_id")
	if scheduleID == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Missing schedule_id parameter",
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

	// Fetch courses from the database or service
	courses, err := scheduler.GetCoursesForSchedule(id)
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
		"Courses":      courses,
		"Instructors":  instructors,
		"Rooms":        rooms,
		"TimeSlots":    timeSlots,
		"CSRFToken":    csrf.GetToken(c),
	}

	c.HTML(http.StatusOK, "courses", data)
}

// SaveCoursesGin handles POST requests to save course changes
func (scheduler *wmu_scheduler) SaveCoursesGin(c *gin.Context) {

	_, err := scheduler.getCurrentUser(c)
	if err != nil {
		log.Printf("Authentication error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse the courses JSON data from the form
	coursesJSON := c.PostForm("courses")
	if coursesJSON == "" {
		log.Printf("No courses data provided in form")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No courses data provided"})
		return
	}

	// Parse JSON into course data structures
	var courses []map[string]interface{}
	err = json.Unmarshal([]byte(coursesJSON), &courses)
	if err != nil {
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
		scheduleID := getIntFromInterface(courseData["schedule_id"])
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

		err = scheduler.AddOrUpdateCourse(crn, section, scheduleID, courseNumber, title, minCredits, maxCredits, minContact, maxContact, cap, approval, lab, instructorID, timeslotID, roomID, mode, status, comment)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to update course ID %d: %v", id, err))
			continue
		}
		successCount++
	}

	// Respond with summary
	if len(errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%d courses updated, %d errors", successCount, len(errors)),
			"errors":  errors,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("All %d courses updated successfully", successCount),
		})
	}
}

func (scheduler *wmu_scheduler) RenderHomePageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

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

// RenderRoomsPageGin renders the rooms page
func (scheduler *wmu_scheduler) RenderRoomsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	rooms, err := scheduler.GetAllRooms()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "rooms.html", gin.H{
			"Error": "Error loading rooms: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Rooms": rooms,
		"User":  user,
	}

	c.HTML(http.StatusOK, "rooms.html", data)
}

// RenderTimeslotsPageGin renders the timeslots page
func (scheduler *wmu_scheduler) RenderTimeslotsPageGin(c *gin.Context) {
	user, err := scheduler.getCurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get all time slots (you may need to implement this method)
	timeslots, err := scheduler.GetAllTimeSlots()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "timeslots.html", gin.H{
			"Error": "Error loading timeslots: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Timeslots": timeslots,
		"User":      user,
	}

	c.HTML(http.StatusOK, "timeslots.html", data)
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
		c.HTML(http.StatusInternalServerError, "instructors.html", gin.H{
			"Error": "Error loading instructors: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Instructors": instructors,
		"User":        user,
	}

	c.HTML(http.StatusOK, "instructors.html", data)
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
	}

	c.HTML(http.StatusOK, "departments.html", data)
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

	prefixes, err := scheduler.GetAllPrefixes()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "prefixes.html", gin.H{
			"Error": "Error loading prefixes: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Prefixes": prefixes,
		"User":     user,
	}

	c.HTML(http.StatusOK, "prefixes.html", data)
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

	users, err := scheduler.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "users.html", gin.H{
			"Error": "Error loading users: " + err.Error(),
			"User":  user,
		})
		return
	}

	data := gin.H{
		"Users": users,
		"User":  user,
	}

	c.HTML(http.StatusOK, "users.html", data)
}

func (scheduler *wmu_scheduler) DeleteScheduleGin(c *gin.Context) {
	// Get schedule ID from form (select box)
	scheduleID := c.Request.URL.Query().Get("schedule_id")
	if scheduleID == "" {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"Error": "No schedule selected for deletion.",
		})
		return
	}
	// Convert scheduleID to integer
	id, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "home.html", gin.H{
			"Error": "Invalid schedule ID.",
		})
		return
	}

	// Attempt to delete the schedule using the database method
	err = scheduler.DeleteSchedule(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "home.html", gin.H{
			"Error": "Failed to delete schedule: " + err.Error(),
		})
		return
	}

	// Redirect to home page with success message
	c.Redirect(http.StatusFound, "/scheduler?success=Schedule deleted successfully")
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
			log.Printf("Error importing course CRN %s: %v", courseData.CRN, err)
			errorCount++
		} else {
			importedCount++
		}
	}

	log.Printf("Import completed: %d courses imported, %d errors", importedCount, errorCount)
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
			log.Printf("Warning: Could not create time slot for %s %s: %v", data.Days, data.Time, err)
			timeSlotID = -1 // This will be converted to NULL
		}
	}

	// Parse room
	roomID := -1
	if data.Location != "" {
		var err error
		roomID, err = scheduler.findOrCreateRoom(data.Location)
		if err != nil {
			log.Printf("Warning: Could not create room for %s: %v", data.Location, err)
			roomID = -1 // This will be converted to NULL
		}
	}

	// Parse instructor
	instructorID := -1
	if data.PrimaryInstructor != "" {
		var err error
		instructorID, err = scheduler.findOrCreateInstructor(data.PrimaryInstructor)
		if err != nil {
			log.Printf("Warning: Could not create instructor for %s: %v", data.PrimaryInstructor, err)
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

	err = scheduler.AddOrUpdateCourse(crn, sectionInt, schedule.ID, courseNum, data.Title,
		minCredits, maxCredits, minContactHours, maxContactHours, capacity, appr, lab, instructorID, timeSlotID,
		roomID, data.MeetingType, "Scheduled", data.Comment)

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

// Helper functions for safe type conversion from interface{}
func getStringFromInterface(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	// Fallback to string representation
	return fmt.Sprintf("%v", value)
}

func getIntFromInterface(value interface{}) int {
	if value == nil {
		return 0
	}

	// Try direct int conversion
	if intVal, ok := value.(int); ok {
		return intVal
	}

	// Try float64 (common from JSON)
	if floatVal, ok := value.(float64); ok {
		return int(floatVal)
	}

	// Try string conversion
	if strVal, ok := value.(string); ok {
		if intVal, err := strconv.Atoi(strVal); err == nil {
			return intVal
		}
	}

	// Fallback: convert to string then to int
	strVal := fmt.Sprintf("%.0f", value)
	if intVal, err := strconv.Atoi(strVal); err == nil {
		return intVal
	}

	return 0
}
