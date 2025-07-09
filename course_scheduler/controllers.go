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

func (scheduler *wmu_scheduler) CheckSession(w http.ResponseWriter, r *http.Request) (bool, error) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/scheduler/signup", http.StatusFound)
		return false, err
	}

	username := cookie.Value
	return scheduler.GetUserLoggedInStatus(username)
}
