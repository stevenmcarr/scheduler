package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if the email format is valid.
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

// connectDB connects to the MySQL database using environment variables
func connectDB() (*sql.DB, error) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}

	if dbUser == "" || dbPassword == "" || dbName == "" {
		return nil, fmt.Errorf("database credentials not found in environment variables")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}

// addUser inserts a new user into the database with bcrypt password hashing
func addUser(db *sql.DB, username, email, password string, administrator bool) error {
	// Validate email
	if !ValidateEmail(email) {
		return fmt.Errorf("invalid email address")
	}

	// Validate password
	if !ValidatePassword(password) {
		return fmt.Errorf("password does not meet requirements: must be at least 15 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	// Hash the password using bcrypt with default cost (same as application)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// Check if user already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking for existing user: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("user with username '%s' or email '%s' already exists", username, email)
	}

	// Insert the new user with administrator flag
	_, err = db.Exec("INSERT INTO users (username, email, password, administrator) VALUES (?, ?, ?, ?)", username, email, hashedPassword, administrator)
	if err != nil {
		return fmt.Errorf("error inserting user into database: %v", err)
	}

	return nil
}

// readPassword reads a password from stdin without echoing it
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input
	return string(password), err
}

// readInput reads a line from stdin
func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func main() {
	fmt.Println("WMU Scheduler - Add User Script")
	fmt.Println("==============================")

	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Printf("Connected to database: %s\n\n", os.Getenv("DB_NAME"))

	// Get user input
	username, err := readInput("Enter username: ")
	if err != nil {
		log.Fatalf("Error reading username: %v", err)
	}

	email, err := readInput("Enter email: ")
	if err != nil {
		log.Fatalf("Error reading email: %v", err)
	}

	password, err := readPassword("Enter password: ")
	if err != nil {
		log.Fatalf("Error reading password: %v", err)
	}

	confirmPassword, err := readPassword("Confirm password: ")
	if err != nil {
		log.Fatalf("Error reading password confirmation: %v", err)
	}

	adminInput, err := readInput("Make user administrator? (y/N): ")
	if err != nil {
		log.Fatalf("Error reading administrator input: %v", err)
	}
	administrator := strings.ToLower(adminInput) == "y" || strings.ToLower(adminInput) == "yes"

	// Validate input
	if username == "" || email == "" || password == "" {
		log.Fatal("Username, email, and password are required")
	}

	if password != confirmPassword {
		log.Fatal("Passwords do not match")
	}

	// Add the user
	err = addUser(db, username, email, password, administrator)
	if err != nil {
		log.Fatalf("Failed to add user: %v", err)
	}

	adminStatus := "regular user"
	if administrator {
		adminStatus = "administrator"
	}
	fmt.Printf("\nUser '%s' added successfully as %s!\n", username, adminStatus)
}
