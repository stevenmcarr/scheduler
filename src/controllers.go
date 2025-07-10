package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// FormData represents data passed to templates
type FormData struct {
	Error   string
	Success string
	Values  map[string]string
}

func (scheduler *wmu_scheduler) ShowSignupForm(w http.ResponseWriter, r *http.Request) {
	scheduler.renderSignupForm(w, "", "", nil)
}

func (scheduler *wmu_scheduler) renderSignupForm(w http.ResponseWriter, errorMsg, successMsg string, values map[string]string) {
	tmplPath := filepath.Join("templates", "signup.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading signup form", http.StatusInternalServerError)
		return
	}

	data := FormData{
		Error:   errorMsg,
		Success: successMsg,
		Values:  values,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering signup form", http.StatusInternalServerError)
	}
}

func (scheduler *wmu_scheduler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	// Check for success message from URL parameters
	successMsg := r.URL.Query().Get("success")
	scheduler.renderLoginForm(w, "", successMsg, nil)
}

func (scheduler *wmu_scheduler) renderLoginForm(w http.ResponseWriter, errorMsg, successMsg string, values map[string]string) {
	tmplPath := filepath.Join("templates", "login.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading login form", http.StatusInternalServerError)
		return
	}

	data := FormData{
		Error:   errorMsg,
		Success: successMsg,
		Values:  values,
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
	cookie, _ := r.Cookie("session")
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
	cookie, err := r.Cookie("session")
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

func (scheduler *wmu_scheduler) LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Preserve form values for re-display
	values := map[string]string{
		"username": username,
	}

	if username == "" || password == "" {
		scheduler.renderLoginForm(w, "Username and password are required", "", values)
		return
	}

	loggedIn, err := scheduler.AuthenticateUser(username, password)
	if err != nil {
		scheduler.renderLoginForm(w, "Error: "+err.Error(), "", values)
		return
	}

	if !loggedIn {
		scheduler.renderLoginForm(w, "Invalid username or password", "", values)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: username,
	})

	err = scheduler.SetUserLoggedInStatus(username, true)

	if err != nil {
		scheduler.renderLoginForm(w, "Error updating login status: "+err.Error(), "", values)
		return
	}

	http.Redirect(w, r, "/scheduler", http.StatusFound)
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
		scheduler.renderSignupForm(w, "All fields are required", "", values)
		return
	}

	// Use email as username since that's what the database expects
	err := scheduler.AddUser(username, email, password)
	if err != nil {
		scheduler.renderSignupForm(w, err.Error(), "", values)
		return
	}

	// Show success message and redirect after a moment, or just redirect
	http.Redirect(w, r, "/scheduler/login?success=Account created successfully", http.StatusFound)
}
