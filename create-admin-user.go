package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get database connection parameters
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	if dbPort == "" {
		dbPort = "3306"
	}

	// Connect to database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// User details
	username := "steve.carr"
	email := "steve.carr@wmich.edu"
	password := "Denture-Subsoil-Overrun1-Ambiance-Unnamed-Outbound"

	// Hash the password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Insert user with administrator privileges
	query := `INSERT INTO users (username, email, password, administrator, is_logged_in) 
			  VALUES (?, ?, ?, 1, 0)
			  ON DUPLICATE KEY UPDATE 
			  password = VALUES(password), 
			  administrator = 1`

	_, err = db.Exec(query, username, email, hashed)
	if err != nil {
		log.Fatal("Failed to insert/update user:", err)
	}

	fmt.Printf("✅ Successfully created/updated administrator user: %s\n", username)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Administrator: Yes\n")
	fmt.Printf("   Password: [ENCRYPTED]\n")

	// Verify the user was created
	var userID int
	var isAdmin bool
	err = db.QueryRow("SELECT id, administrator FROM users WHERE username = ?", username).Scan(&userID, &isAdmin)
	if err != nil {
		log.Fatal("Failed to verify user creation:", err)
	}

	fmt.Printf("   User ID: %d\n", userID)
	fmt.Printf("   Admin status confirmed: %v\n", isAdmin)
}
