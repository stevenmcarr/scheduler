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
	Username      string
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
