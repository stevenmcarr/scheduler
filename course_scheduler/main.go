package main

import (
	"database/sql"
	"fmt"
)

type wmu_scheduler struct {
	database *sql.DB
}

func main() {
	// Connect to MySQL server
	database, err := ConnectMySQL("wmu_cs", "1h0ck3y$", "wmu_schedules")
	if err != nil {
		panic(err)
	}

	scheduler := &wmu_scheduler{
		database: database,
	}

	// Create Gin router with default middleware (logger and recovery)
	r := scheduler.router()

	defer database.Close()
	// Start server on port 8080
	r.Run(":8080")
}

func ConnectMySQL(user, password, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", user, password, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
