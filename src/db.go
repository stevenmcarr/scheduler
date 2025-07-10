package main

import (
	"database/sql"
	"errors"
	"regexp"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system.
type User struct {
	ID            int
	Username      string
	Email         string
	Password      string
	IsLoggedIn    bool
	Administrator bool
}

// Email regex (simple version)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the email is valid.
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePassword checks if the password meets the requirements.
func ValidatePassword(password string) bool {
	if len(password) < 15 {
		return false
	}
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// AddUser inserts a new user into the users table.
func (scheduler *wmu_scheduler) AddUser(username, email, password string) error {
	if !ValidateEmail(email) {
		return errors.New("invalid email address")
	}
	if !ValidatePassword(password) {
		return errors.New("password does not meet requirements: must be at least 15 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = scheduler.database.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, hashed)
	return err
}

// DeleteUser removes a user from the users table.
func (scheduler *wmu_scheduler) DeleteUser(username string) error {
	_, err := scheduler.database.Exec("DELETE FROM users WHERE username = ?", username)
	return err
}

// UpdateUserPassword updates the password for a user.
func (scheduler *wmu_scheduler) UpdateUserPassword(username, newPassword string) error {
	if !ValidatePassword(newPassword) {
		return errors.New("password does not meet requirements")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = scheduler.database.Exec("UPDATE users SET password = ? WHERE username = ?", hashed, username)
	return err
}

func (scheduler *wmu_scheduler) AuthenticateUser(usernameOrEmail, password string) (bool, error) {
	var hashedPassword string
	// Allow login with either username or email
	err := scheduler.database.QueryRow("SELECT password FROM users WHERE username = ? OR email = ?", usernameOrEmail, usernameOrEmail).Scan(&hashedPassword)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil, nil
}

func (scheduler *wmu_scheduler) GetUserLoggedInStatus(usernameOrEmail string) (bool, error) {
	var isLoggedIn bool
	err := scheduler.database.QueryRow("SELECT is_logged_in FROM users WHERE username = ? OR email = ?", usernameOrEmail, usernameOrEmail).Scan(&isLoggedIn)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isLoggedIn, nil
}

func (scheduler *wmu_scheduler) SetUserLoggedInStatus(usernameOrEmail string, isLoggedIn bool) error {
	_, err := scheduler.database.Exec("UPDATE users SET is_logged_in = ? WHERE username = ? OR email = ?", isLoggedIn, usernameOrEmail, usernameOrEmail)
	return err
}

func (scheduler *wmu_scheduler) GetUserByUsername(username string) (*User, error) {
	var user User
	err := scheduler.database.QueryRow("SELECT id, username, email, is_logged_in, administrator FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &user.IsLoggedIn, &user.Administrator)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (scheduler *wmu_scheduler) GetUserByEmail(email string) (*User, error) {
	var user User
	err := scheduler.database.QueryRow("SELECT id, username, email, is_logged_in, administrator FROM users WHERE email = ?", email).Scan(&user.ID, &user.Username, &user.Email, &user.IsLoggedIn, &user.Administrator)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

type Schedule struct {
	ID         int
	Term       string
	Year       int
	Department string
	Prefix     string
	Created    string
}

func (scheduler *wmu_scheduler) AddSchedule(term string, year int, prefix string) error {
	_, err := scheduler.database.Exec("INSERT INTO schedules (term, year, prefix) VALUES (?, ?, ?)", term, year, prefix)
	return err
}

func (scheduler *wmu_scheduler) DeleteSchedule(term string, year int, prefix string) error {
	_, err := scheduler.database.Exec("DELETE FROM schedules WHERE term = ? AND year = ? AND prefix = ?", term, year, prefix)
	return err
}

func (scheduler *wmu_scheduler) GetSchedule(term string, year int, prefix string) (*Schedule, error) {
	var schedule Schedule
	err := scheduler.database.QueryRow("SELECT id, term, year, prefix FROM schedules WHERE term = ? AND year = ? AND prefix = ?", term, year, prefix).Scan(&schedule.ID, &schedule.Term, &schedule.Year, &schedule.Prefix)
	if err == sql.ErrNoRows {
		return nil, nil // Schedule not found
	}
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (scheduler *wmu_scheduler) GetAllSchedules() ([]Schedule, error) {
	rows, err := scheduler.database.Query(`
		SELECT s.id, s.term, s.year, p.prefix, d.name, s.created_at 
		FROM schedules s
		JOIN prefixes p ON s.prefix_id = p.id
		JOIN departments d ON s.department_id = d.id
		ORDER BY s.year DESC, s.term, d.name, p.prefix
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		var schedule Schedule
		if err := rows.Scan(&schedule.ID, &schedule.Term, &schedule.Year, &schedule.Prefix, &schedule.Department, &schedule.Created); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (scheduler *wmu_scheduler) UpdateSchedule(term string, year int, prefix string) error {
	_, err := scheduler.database.Exec("UPDATE schedules SET term = ?, year = ?, prefix = ? WHERE term = ? AND year = ? AND prefix = ?", term, year, prefix, term, year, prefix)
	return err
}

func (scheduler *wmu_scheduler) GetScheduleByID(id int) (*Schedule, error) {
	var schedule Schedule
	err := scheduler.database.QueryRow("SELECT id, term, year, prefix FROM schedules WHERE id = ?", id).Scan(&schedule.ID, &schedule.Term, &schedule.Year, &schedule.Prefix)
	if err == sql.ErrNoRows {
		return nil, nil // Schedule not found
	}
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (scheduler *wmu_scheduler) GetSchedulesByTerm(term string) ([]Schedule, error) {
	rows, err := scheduler.database.Query("SELECT id, term, year, prefix FROM schedules WHERE term = ?", term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		var schedule Schedule
		if err := rows.Scan(&schedule.ID, &schedule.Term, &schedule.Year, &schedule.Prefix); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}

func (scheduler *wmu_scheduler) GetSchedulesByYear(year int) ([]Schedule, error) {
	rows, err := scheduler.database.Query("SELECT id, term, prefix FROM schedules WHERE year = ?", year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []Schedule
	for rows.Next() {
		var schedule Schedule
		if err := rows.Scan(&schedule.ID, &schedule.Term, &schedule.Prefix); err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}
	return schedules, nil
}
