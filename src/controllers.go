package main

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func (scheduler *wmu_scheduler) ShowSignupForm(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "signup.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading signup form", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering signup form", http.StatusInternalServerError)
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

func (scheduler *wmu_scheduler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "login.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Error loading login form", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering login form", http.StatusInternalServerError)
	}
}

func (scheduler *wmu_scheduler) LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	loggedIn, err := scheduler.AuthenticateUser(username, password)
	if err != nil || !loggedIn {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: username,
	})
	http.Redirect(w, r, "/scheduler", http.StatusFound)
}
