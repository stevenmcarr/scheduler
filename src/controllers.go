package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
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

func (scheduler *wmu_scheduler) RenderHomePage(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	if !scheduler.CheckSession(w, r) {
		http.Redirect(w, r, "/scheduler/login", http.StatusFound)
		return
	}

	// Get current user for navbar
	cookie, _ := r.Cookie("user")
	username := cookie.Value
	user, err := scheduler.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Parse both navbar and home templates
	tmplPath := filepath.Join("templates", "home.html")
	navbarPath := filepath.Join("templates", "navbar.html")
	tmpl, err := template.ParseFiles(tmplPath, navbarPath)
	if err != nil {
		http.Error(w, "Error loading home page", http.StatusInternalServerError)
		return
	}

	schedules, err := scheduler.GetAllSchedules()
	if err != nil {
		http.Error(w, "Error fetching schedules", http.StatusInternalServerError)
		return
	}

	// Construct the data structure expected by the template
	data := struct {
		User      *User
		Schedules []Schedule
	}{
		User:      user,
		Schedules: schedules,
	}

	err = tmpl.ExecuteTemplate(w, "WMU Course Scheduler", data)
	if err != nil {
		http.Error(w, "Error rendering home page", http.StatusInternalServerError)
	}
}

func (scheduler *wmu_scheduler) CheckSession(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("user")
	if err != nil || cookie.Value == "" {
		return false
	}

	username := cookie.Value
	loggedIn, err := scheduler.GetUserLoggedInStatus(username)
	if err != nil {
		return false
	}
	return loggedIn
}

func (scheduler *wmu_scheduler) SignupUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Preserve form values for re-display
	values := map[string]string{
		"username": username,
		"email":    email,
	}

	if username == "" || email == "" || password == "" {
		scheduler.renderSignupForm(w, r, "All fields are required", "", values)
		return
	}

	// Use email as username since that's what the database expects
	err := scheduler.AddUser(username, email, password)
	if err != nil {
		scheduler.renderSignupForm(w, r, err.Error(), "", values)
		return
	}

	// Show success message and redirect after a moment, or just redirect
	http.Redirect(w, r, "/scheduler/login?success=Account created successfully", http.StatusFound)
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

	// Set session cookie
	c.SetCookie("user", username, 3600, "/", "", false, true)

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
	if !scheduler.CheckSession(c.Writer, c.Request) {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Check if user is logged in
	// Get schedule_id from the URL query parameters
	scheduleID := c.Request.URL.Query().Get("schedule_id")
	if scheduleID == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Missing schedule_id parameter",
		})
		return
	}

	id, err := strconv.Atoi(scheduleID)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"Error": "Invalid schedule_id parameter",
		})
		return
	}

	// Fetch courses from the database or service
	courses, err := scheduler.GetAllCourses(id)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching courses",
		})
		return
	}

	// Get current user for navbar
	cookie, err := c.Request.Cookie("user")
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching user session",
		})
		return
	}
	username := cookie.Value
	user, err := scheduler.GetUserByUsername(username)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Error fetching user data",
		})
		return
	}

	data := gin.H{
		"User":      user,
		"Courses":   courses,
		"CSRFToken": csrf.GetToken(c),
	}

	c.HTML(http.StatusOK, "courses.html", data)
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
	// Check if user is logged in
	user, err := c.Cookie("user")
	if err != nil {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	if user == "" {
		c.Redirect(http.StatusFound, "/scheduler/login")
		return
	}

	// Get current user
	currentUser, err := scheduler.GetUserByUsername(user)
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
	user, err := c.Cookie("user")
	if err != nil {
		return nil, err
	}

	if user == "" {
		return nil, fmt.Errorf("no session")
	}

	return scheduler.GetUserByUsername(user)
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
