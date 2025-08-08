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
	var multiWriter io.Writer

	if err != nil {
		// If we can't open the log file, fall back to stdout only
		multiWriter = os.Stdout
		// Still create the AppLogger, but warn about the fallback
		AppLogger = &Logger{
			Logger: log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile),
		}
		// Use the newly created AppLogger to log the warning
		AppLogger.LogWarning(fmt.Sprintf("Failed to open log file, using stdout only: %v", err))
	} else {
		// Create multi-writer to write to both file and stdout
		multiWriter = io.MultiWriter(os.Stdout, logFile)
		// Create custom logger
		AppLogger = &Logger{
			Logger: log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile),
		}
	}

	// Set up the standard log package
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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
	initLogger() // Now always succeeds, may fallback to stdout-only logging

	AppLogger.LogInfo("Starting WMU Course Scheduler...")

	// Load environment variables from .env file
	// Check if a custom env file is specified
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env" // default
	}

	// Try multiple paths for .env file - prioritize local development environment
	// Order: 1) current directory, 2) custom env file, 3) parent directory, 4) production path
	envPaths := []string{".env"}
	if envFile != ".env" {
		envPaths = append(envPaths, envFile)
	}
	envPaths = append(envPaths, "../.env", "/var/www/html/scheduler/.env")
	var envErr error
	AppLogger.LogInfo(fmt.Sprintf("Searching for .env file in order: %v", envPaths))
	for _, path := range envPaths {
		envErr = godotenv.Load(path)
		if envErr == nil {
			AppLogger.LogInfo(fmt.Sprintf("✅ Successfully loaded environment from: %s", path))
			break
		} else {
			AppLogger.LogInfo(fmt.Sprintf("❌ Could not load: %s (%v)", path, envErr))
		}
	}
	if envErr != nil {
		AppLogger.LogError("Error loading .env file from all paths", envErr)
		// Use direct os.Exit instead of log.Fatal to avoid duplicate logging
		os.Exit(1)
	}

	// Get database connection parameters from environment variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	serverPort := os.Getenv("SERVER_PORT")

	if dbUser == "" || dbPassword == "" || dbName == "" {
		AppLogger.LogError("Missing required database environment variables", nil)
		os.Exit(1)
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
		os.Exit(1)
	}

	AppLogger.LogInfo("Database connection established successfully")

	scheduler := &wmu_scheduler{
		database: database,
	}

	// Get TLS configuration
	tlsEnabled := os.Getenv("TLS_ENABLED")
	tlsCertFile := os.Getenv("TLS_CERT_FILE")
	tlsKeyFile := os.Getenv("TLS_KEY_FILE")

	// Create Gin router with default middleware (logger and recovery)
	r := scheduler.router()

	defer func() {
		AppLogger.LogInfo("Closing database connection...")
		database.Close()
	}()

	// Start server with TLS if enabled
	if tlsEnabled == "true" {
		var certFile, keyFile string
		var tlsSource string

		// Check for Let's Encrypt certificates first
		if tlsCertFile != "" && tlsKeyFile != "" {
			// Use environment-specified certificates
			if _, err := os.Stat(tlsCertFile); err == nil {
				if _, err := os.Stat(tlsKeyFile); err == nil {
					certFile = tlsCertFile
					keyFile = tlsKeyFile
					tlsSource = "environment configuration"
				}
			}
		}

		// Fall back to self-signed certificate
		if certFile == "" {
			selfSignedCert := "certs/server.crt"
			selfSignedKey := "certs/server.key"
			if _, err := os.Stat(selfSignedCert); err == nil {
				if _, err := os.Stat(selfSignedKey); err == nil {
					certFile = selfSignedCert
					keyFile = selfSignedKey
					tlsSource = "self-signed"
				}
			}
		}

		if certFile != "" && keyFile != "" {
			AppLogger.LogInfo(fmt.Sprintf("Starting HTTPS server on port %s...", serverPort))
			AppLogger.LogInfo(fmt.Sprintf("Using %s TLS certificate: %s", tlsSource, certFile))
			err = r.RunTLS(":"+serverPort, certFile, keyFile)
		} else {
			AppLogger.LogWarning("TLS enabled but no valid certificates found, falling back to HTTP")
			AppLogger.LogInfo(fmt.Sprintf("Starting HTTP server on port %s...", serverPort))
			err = r.Run(":" + serverPort)
		}
	} else {
		AppLogger.LogInfo(fmt.Sprintf("Starting HTTP server on port %s...", serverPort))
		err = r.Run(":" + serverPort)
	}

	if err != nil {
		AppLogger.LogError("Failed to start server", err)
		os.Exit(1)
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
