package main

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func (scheduler *wmu_scheduler) AddOrGetSchedule(term string, year int, prefix string) (*Schedule, error) {
	// Check if schedule already exists
	var scheduleID int
	var created string
	var prefixID int
	var departmentName string
	var departmentID int
	err := scheduler.database.QueryRow(`
		SELECT s.id, s.prefix_id, s.created_at, s.department_id, d.name
		FROM schedules s
		JOIN departments d ON s.department_id = d.id
		WHERE s.term = ? AND s.year = ?
	`, term, year).Scan(&scheduleID, &prefixID, &created, &departmentID, &departmentName)

	if err == nil {
		return &Schedule{
			ID:         scheduleID,
			Term:       term,
			Year:       year,
			Department: departmentName,
			Prefix:     prefix,
			Created:    created,
		}, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	prefixID, err = strconv.Atoi(prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix id: %v", err)
	}
	// Get department_id and prefix_id from prefixes table/
	err = scheduler.database.QueryRow("SELECT id, department_id FROM prefixes WHERE id = ?", prefixID).Scan(&prefixID, &departmentID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prefix '%s' not found", prefix)
	}
	if err != nil {
		return nil, err
	}
	// Insert new schedule
	response, err := scheduler.database.Exec(
		"INSERT INTO schedules (term, year, department_id, prefix_id) VALUES (?, ?, ?, ?)",
		term, year, departmentID, prefixID,
	)
	if err != nil {
		return nil, err
	}
	id, err := response.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get created_at for the new schedule
	err = scheduler.database.QueryRow("SELECT created_at FROM schedules WHERE id = ?", id).Scan(&created)
	if err != nil {
		return nil, err
	}

	return &Schedule{
		ID:         int(id),
		Term:       term,
		Year:       year,
		Department: departmentName,
		Prefix:     prefix,
		Created:    created,
	}, nil
}

func (scheduler *wmu_scheduler) DeleteSchedule(id int) error {
	// Begin a transaction to ensure both operations succeed or fail together
	tx, err := scheduler.database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // This will be ignored if the transaction is committed

	// First, delete all courses associated with this schedule
	_, err = tx.Exec("DELETE FROM courses WHERE schedule_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete courses for schedule %d: %v", id, err)
	}

	// Then delete the schedule itself
	result, err := tx.Exec("DELETE FROM schedules WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule %d: %v", id, err)
	}

	// Check if the schedule actually existed
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("schedule with id %d not found", id)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (scheduler *wmu_scheduler) GetSchedule(term string, year int, prefix string) (*Schedule, error) {
	var schedule Schedule
	err := scheduler.database.QueryRow(`
		SELECT s.id, s.term, s.year, d.name, p.prefix, s.created_at
		FROM schedules s
		JOIN prefixes p ON s.prefix_id = p.id
		JOIN departments d ON s.department_id = d.id
		WHERE s.term = ? AND s.year = ? AND p.prefix = ?
	`, term, year, prefix).Scan(&schedule.ID, &schedule.Term, &schedule.Year, &schedule.Department, &schedule.Prefix, &schedule.Created)
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

type Course struct {
	ID           int
	CRN          int
	Section      string
	ScheduleID   int
	Prefix       string
	CourseNumber string
	Title        string
	Credits      int
	Contact      string
	Cap          int
	Approval     bool // Changed from Appr to Approval
	Lab          bool
	InstructorID int
	TimeSlotID   int // New field for timeslot ID
	RoomID       int // New field for room ID
	Mode         string
	Status       string
	Comment      string // New field for comments
}

func (scheduler *wmu_scheduler) GetCoursesForSchedule(scheduleID int) ([]Course, error) {
	rows, err := scheduler.database.Query(`
		SELECT c.id, c.crn, c.section, p.prefix, c.course_number, c.title, c.min_credits, c.max_contact, c.cap, 
		       c.approval = 1 as approval, c.lab = 1 as lab,
			   COALESCE(c.instructor_id, -1) as instructor_id,
			   COALESCE(c.timeslot_id, -1) as timeslot_id,
			   COALESCE(c.room_id, -1) as room_id,
		       c.mode, c.status, c.comment
		FROM courses c
		JOIN schedules s ON c.schedule_id = s.id
		JOIN prefixes p ON s.prefix_id = p.id
		WHERE c.schedule_id = ?
		ORDER BY c.course_number, c.crn, c.section
	`, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var course Course
		course.ScheduleID = scheduleID // Set ScheduleID from the parameter
		if err := rows.Scan(&course.ID, &course.CRN, &course.Section, &course.Prefix, &course.CourseNumber, &course.Title, &course.Credits, &course.Contact, &course.Cap, &course.Approval, &course.Lab, &course.InstructorID, &course.TimeSlotID, &course.RoomID, &course.Mode, &course.Status, &course.Comment); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

type Room struct {
	ID           int
	Building     string
	RoomNumber   string
	Capacity     int
	ComputerLab  bool
	DedicatedLab bool
}

func (scheduler *wmu_scheduler) GetAllRooms() ([]Room, error) {
	rows, err := scheduler.database.Query("SELECT id,building, room_number, capacity, computer_lab, dedicated_lab FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.Building, &room.RoomNumber, &room.Capacity, &room.ComputerLab, &room.DedicatedLab); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

type Prefix struct {
	ID         int
	Prefix     string
	Department string
}

func (scheduler *wmu_scheduler) GetAllPrefixes() ([]Prefix, error) {
	rows, err := scheduler.database.Query(`
		SELECT p.id, p.prefix, d.name
		FROM prefixes p
		JOIN departments d ON p.department_id = d.id
		ORDER BY p.prefix
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefixes []Prefix
	for rows.Next() {
		var prefix Prefix
		if err := rows.Scan(&prefix.ID, &prefix.Prefix, &prefix.Department); err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func (scheduler *wmu_scheduler) GetPrefixForSchedule(scheduleID int) (*Prefix, error) {
	var prefix Prefix
	err := scheduler.database.QueryRow(`
		SELECT p.id, p.prefix, d.name
		FROM schedules s
		JOIN prefixes p ON s.prefix_id = p.id
		JOIN departments d ON p.department_id = d.id
		WHERE s.id = ?
	`, scheduleID).Scan(&prefix.ID, &prefix.Prefix, &prefix.Department)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	return &prefix, nil
}

type TimeSlot struct {
	ID        int
	StartTime string
	EndTime   string
	Days      string
	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
}

// GetAllTimeSlots retrieves all time slots from the database
func (scheduler *wmu_scheduler) GetAllTimeSlots() ([]TimeSlot, error) {
	query := "SELECT id, start_time, end_time, M, T, W, R, F FROM time_slots ORDER BY start_time, end_time, M, T, W, R, F"
	rows, err := scheduler.database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeslots []TimeSlot
	for rows.Next() {
		var timeslot TimeSlot
		err := rows.Scan(&timeslot.ID, &timeslot.StartTime, &timeslot.EndTime, &timeslot.Monday, &timeslot.Tuesday, &timeslot.Wednesday, &timeslot.Thursday, &timeslot.Friday)
		if err != nil {
			return nil, err
		}
		timeslot.Days = "" // Initialize Days field
		// Set Days field based on boolean values
		if timeslot.Monday {
			timeslot.Days += "M"
		}
		if timeslot.Tuesday {
			timeslot.Days += "T"
		}
		if timeslot.Wednesday {
			timeslot.Days += "W"
		}
		if timeslot.Thursday {
			timeslot.Days += "R"
		}
		if timeslot.Friday {
			timeslot.Days += "F"
		}
		timeslots = append(timeslots, timeslot)
	}

	return timeslots, nil
}

type Instructor struct {
	ID         int
	LastName   string
	FirstName  string
	Department string
	Status     string
}

// GetAllInstructors retrieves all instructors from the database
func (scheduler *wmu_scheduler) GetAllInstructors() ([]Instructor, error) {
	query := `
		SELECT i.id, i.last_name, i.first_name, d.name, i.status
		FROM instructors i
		JOIN departments d ON i.department_id = d.id
		ORDER BY i.last_name, i.first_name
	`
	rows, err := scheduler.database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instructors []Instructor
	for rows.Next() {
		var instructor Instructor
		err := rows.Scan(&instructor.ID, &instructor.LastName, &instructor.FirstName, &instructor.Department, &instructor.Status)
		if err != nil {
			return nil, err
		}
		instructors = append(instructors, instructor)
	}

	return instructors, nil
}

type Department struct {
	ID   int
	Name string
}

// GetAllDepartments retrieves all departments from the database
func (scheduler *wmu_scheduler) GetAllDepartments() ([]Department, error) {
	query := "SELECT id, name FROM departments ORDER BY name"
	rows, err := scheduler.database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []Department
	for rows.Next() {
		var department Department
		err := rows.Scan(&department.ID, &department.Name)
		if err != nil {
			return nil, err
		}
		departments = append(departments, department)
	}

	return departments, nil
}

// GetAllUsers retrieves all users from the database
func (scheduler *wmu_scheduler) GetAllUsers() ([]User, error) {
	query := "SELECT id, username, email, is_logged_in, administrator FROM users ORDER BY username"
	rows, err := scheduler.database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsLoggedIn, &user.Administrator)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (scheduler *wmu_scheduler) GetPrefix(prefix string) (*Prefix, error) {
	var p Prefix
	err := scheduler.database.QueryRow(`
		SELECT p.id, p.prefix, d.name
		JOIN departments d ON p.department_id = d.id
		FROM prefixes p WHERE p.id = ?`,
		prefix).Scan(&p.ID, &p.Prefix, &p.Department)
	if err == sql.ErrNoRows {
		return nil, nil // Prefix not found
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (scheduler *wmu_scheduler) AddCourse(
	crn int,
	section int,
	courseNumber int,
	title string,
	minCredits int,
	maxCredits int,
	minContact int,
	maxContact int,
	cap int,
	approval bool,
	lab bool,
	instructorID int,
	timeslotID int,
	roomID int,
	mode string,
	status string,
	comment string,
	scheduleID int,
) error {
	// Use nil for MySQL NULL if any of the IDs are -1
	var instructorVal, timeslotVal, roomVal interface{}
	if instructorID == -1 {
		instructorVal = nil
	} else {
		instructorVal = instructorID
	}
	if timeslotID == -1 {
		timeslotVal = nil
	} else {
		timeslotVal = timeslotID
	}
	if roomID == -1 {
		roomVal = nil
	} else {
		roomVal = roomID
	}

	_, err := scheduler.database.Exec(`
		INSERT INTO courses (
			crn, section, schedule_id, course_number, title, min_credits, max_contact, cap, approval, lab, instructor_id, timeslot_id, room_id, mode, status, comment
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, crn, section, scheduleID, courseNumber, title, minCredits, maxCredits, minContact, maxContact, cap, approval, lab, instructorVal, timeslotVal, roomVal, mode, status, comment)
	return err
}

func (scheduler *wmu_scheduler) AddOrUpdateCourse(
	crn int,
	section int,
	courseNumber int,
	title string,
	minCredits int,
	maxCredits int,
	minContactHours int,
	maxContactHours int,
	cap int,
	appr int,
	lab int,
	instructorID int,
	timeslotID int,
	roomID int,
	mode string,
	status string,
	comment string,
	scheduleID int,
) error {
	// Try to update first
	var result sql.Result
	var err error
	// Use nil for MySQL NULL if any of the IDs are -1
	var instructorVal, timeslotVal, roomVal interface{}
	if instructorID == -1 {
		instructorVal = nil
	} else {
		instructorVal = instructorID
	}
	if timeslotID == -1 {
		timeslotVal = nil
	} else {
		timeslotVal = timeslotID
	}
	if roomID == -1 {
		roomVal = nil
	} else {
		roomVal = roomID
	}

	result, err = scheduler.database.Exec(`
		UPDATE courses SET
			section = ?, course_number = ?, title = ?, min_credits = ?, max_credits = ?, min_contact = ?, max_contact = ?, cap = ?, approval = ?, lab = ?, instructor_id = ?, timeslot_id = ?, room_id = ?, mode = ?, status = ?, comment = ?
		WHERE crn = ?
	`, section, courseNumber, title, minCredits, maxCredits, minContactHours, maxContactHours, cap, appr, lab, instructorVal, timeslotVal, roomVal, mode, status, comment, crn)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		return nil // Updated existing course
	}

	// Check if the CRN exists but no update was needed (all values were the same)
	var existingCRN int
	err = scheduler.database.QueryRow("SELECT crn FROM courses WHERE crn = ?", crn).Scan(&existingCRN)
	if err == nil {
		// CRN exists but no update was needed (all values were already the same)
		return nil
	}
	if err != sql.ErrNoRows {
		// Some other error occurred during the check
		return fmt.Errorf("error checking for existing CRN: %v", err)
	}

	// CRN doesn't exist, so insert new course
	_, err = scheduler.database.Exec(`
		INSERT INTO courses (
			crn, section, schedule_id, course_number, title, min_credits, max_credits, min_contact, max_contact, cap, approval, lab, instructor_id, timeslot_id, room_id, mode, status, comment
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, crn, section, scheduleID, courseNumber, title, minCredits, maxCredits, minContactHours, maxContactHours, cap, appr, lab, instructorVal, timeslotVal, roomVal, mode, status, comment)
	return err
}

// Helper functions for finding or creating related entities
func (scheduler *wmu_scheduler) findOrCreateTimeSlot(days, time string) (int, error) {
	// Parse time (e.g., "1130-1245" to start and end times)
	timeParts := strings.Split(time, "-")
	if len(timeParts) != 2 {
		return -1, fmt.Errorf("invalid time format: %s", time)
	}

	startTime, err := parseTime(timeParts[0])
	if err != nil {
		return -1, err
	}

	endTime, err := parseTime(timeParts[1])
	if err != nil {
		return -1, err
	}

	var monday, tuesday, wednesday, thursday, friday bool
	for _, d := range days {
		switch d {
		case 'M':
			monday = true
		case 'T':
			tuesday = true
		case 'W':
			wednesday = true
		case 'R':
			thursday = true
		case 'F':
			friday = true
		}
	}
	// Check if time slot exists
	var id int
	query := "SELECT id FROM time_slots WHERE M = ? AND T = ? AND W = ? AND R = ? AND F = ? AND start_time = ? AND end_time = ?"
	err = scheduler.database.QueryRow(query, monday, tuesday, wednesday, thursday, friday, startTime, endTime).Scan(&id)
	if err == nil {
		return id, nil
	}

	// If not found, create new time slot
	if err != sql.ErrNoRows {
		return -1, fmt.Errorf("error checking for existing time slot: %v", err)
	}

	// Create new time slot
	query = "INSERT INTO time_slots (M, T, W, R, F, start_time, end_time) VALUES (?, ?, ?, ?, ?, ?, ?)"
	result, err := scheduler.database.Exec(query, monday, tuesday, wednesday, thursday, friday, startTime, endTime)
	if err != nil {
		return -1, fmt.Errorf("error creating time slot: %v", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("error getting new time slot ID: %v", err)
	}

	return int(newID), nil
}

func (scheduler *wmu_scheduler) findOrCreateRoom(location string) (int, error) {
	// Parse room (e.g., "D0109 FLOYD" to room number and building)
	parts := strings.Fields(location)
	if len(parts) < 2 {
		return -1, fmt.Errorf("invalid location format: %s", location)
	}

	roomNumber := parts[0]
	building := strings.Join(parts[1:], " ")

	// Check if room exists
	var id int
	query := "SELECT id FROM rooms WHERE room_number = ? AND building = ?"
	err := scheduler.database.QueryRow(query, roomNumber, building).Scan(&id)
	if err == nil {
		return id, nil
	}

	// If not found, create new room
	if err != sql.ErrNoRows {
		return -1, fmt.Errorf("error checking for existing room: %v", err)
	}

	// Create new room
	query = "INSERT INTO rooms (room_number, building, capacity) VALUES (?, ?, ?)"
	result, err := scheduler.database.Exec(query, roomNumber, building, 0) // Default capacity
	if err != nil {
		return -1, fmt.Errorf("error creating room: %v", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("error getting new room ID: %v", err)
	}

	return int(newID), nil
}

func (scheduler *wmu_scheduler) findOrCreateInstructor(name string) (int, error) {
	// Check if instructor exists
	var id int
	query := "SELECT id FROM instructors WHERE name = ?"
	err := scheduler.database.QueryRow(query, name).Scan(&id)
	if err == nil {
		return id, nil
	}

	// If not found, create new instructor
	if err != sql.ErrNoRows {
		return -1, fmt.Errorf("error checking for existing instructor: %v", err)
	}

	// Create new instructor
	query = "INSERT INTO instructors (name, email, department) VALUES (?, ?, ?)"
	result, err := scheduler.database.Exec(query, name, "", "Computer Science") // Default department
	if err != nil {
		return -1, fmt.Errorf("error creating instructor: %v", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("error getting new instructor ID: %v", err)
	}

	return int(newID), nil
}

// UpdateCourseField updates a single field for a course identified by CourseID.
func (scheduler *wmu_scheduler) UpdateCourseField(courseID int, field string, value interface{}) error {
	// Only allow updates to known fields to prevent SQL injection
	allowedFields := map[string]bool{
		"crn":           true,
		"section":       true,
		"prefix":        true,
		"course_number": true,
		"title":         true,
		"min_credits":   true,
		"max_credits":   true,
		"min_contact":   true,
		"max_contact":   true,
		"cap":           true,
		"approval":      true,
		"lab":           true,
		"instructor_id": true,
		"timeslot_id":   true,
		"room_id":       true,
		"mode":          true,
		"status":        true,
		"comment":       true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("field '%s' cannot be updated", field)
	}

	query := fmt.Sprintf("UPDATE courses SET %s = ? WHERE id = ?", field)
	_, err := scheduler.database.Exec(query, value, courseID)
	return err
}

func (scheduler *wmu_scheduler) GetScheduleName(scheduleID int) (string, error) {
	var term string
	var year int
	err := scheduler.database.QueryRow("SELECT term, year FROM schedules WHERE id = ?", scheduleID).Scan(&term, &year)
	if err == sql.ErrNoRows {
		return "", nil // Schedule not found
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %d", term, year), nil
}
