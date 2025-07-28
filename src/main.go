package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Custom logger for error handling
type Logger struct {
	*log.Logger
}

var AppLogger *Logger

// Initialize custom logger
func initLogger() error {
	logFile, err := os.OpenFile("/var/log/scheduler/scheduler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Set up the standard log package
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create custom logger
	AppLogger = &Logger{
		Logger: log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile),
	}

	return nil
}

// LogError logs error messages with timestamp
func (l *Logger) LogError(msg string, err error) {
	if err != nil {
		l.Printf("ERROR: %s - %v", msg, err)
	} else {
		l.Printf("ERROR: %s", msg)
	}
}

// LogInfo logs informational messages
func (l *Logger) LogInfo(msg string) {
	l.Printf("INFO: %s", msg)
}

// LogWarning logs warning messages
func (l *Logger) LogWarning(msg string) {
	l.Printf("WARNING: %s", msg)
}

// LogHTTP logs HTTP request information
func (l *Logger) LogHTTP(method, path, clientIP, userAgent string, statusCode int, latency string) {
	l.Printf("HTTP: %s %s | %d | %s | %s | %s", method, path, statusCode, latency, clientIP, userAgent)
}

type wmu_scheduler struct {
	database *sql.DB
}

func main() {
	// Initialize logging
	err := initLogger()
	if err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		log.Println("Continuing with stdout logging...")
	}

	AppLogger.LogInfo("Starting WMU Course Scheduler...")

	// Load environment variables from .env file
	err = godotenv.Load("../.env")
	if err != nil {
		AppLogger.LogError("Error loading .env file", err)
		log.Fatal("Error loading .env file")
	}

	// Get database connection parameters from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	serverPort := os.Getenv("SERVER_PORT")

	if dbUser == "" || dbPassword == "" || dbName == "" {
		AppLogger.LogError("Missing required database environment variables", nil)
		log.Fatal("Missing required database environment variables")
	}

	if serverPort == "" {
		serverPort = "8080" // default port
		AppLogger.LogWarning("SERVER_PORT not set, using default port 8080")
	}

	AppLogger.LogInfo(fmt.Sprintf("Connecting to database: %s@%s", dbUser, dbName))

	// Connect to MySQL server
	database, err := ConnectMySQL(dbUser, dbPassword, dbName)
	if err != nil {
		AppLogger.LogError("Failed to connect to database", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}

	AppLogger.LogInfo("Database connection established successfully")

	scheduler := &wmu_scheduler{
		database: database,
	}

	// Create Gin router with default middleware (logger and recovery)
	r := scheduler.router()

	defer func() {
		AppLogger.LogInfo("Closing database connection...")
		database.Close()
	}()

	AppLogger.LogInfo(fmt.Sprintf("Starting server on port %s...", serverPort))

	// Start server on configured port
	err = r.Run(":" + serverPort)
	if err != nil {
		AppLogger.LogError("Failed to start server", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}

func ConnectMySQL(user, password, dbname string) (*sql.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	if dbHost == "" {
		dbHost = "127.0.0.1" // default host
	}
	if dbPort == "" {
		dbPort = "3306" // default port
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, dbHost, dbPort, dbname)
	AppLogger.LogInfo(fmt.Sprintf("Attempting to connect to MySQL at %s:%s", dbHost, dbPort))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		AppLogger.LogError("Failed to open MySQL connection", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		AppLogger.LogError("Failed to ping MySQL database", err)
		return nil, err
	}

	return db, nil
}
