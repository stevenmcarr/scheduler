package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ExcelCourseData represents a course row from Excel
type ExcelCourseData struct {
	CRN               string
	CourseID          string
	Section           string
	Status            string
	Title             string
	Link1             string
	Link2             string
	SchedType         string
	Reserved          string
	MinCreditHours    string
	MaxCreditHours    string
	BillingHours      string
	MinContactHours   string
	MaxContactHours   string
	Gradeable         string
	Capacity          string
	WaitlistCap       string
	SpecialApproval   string
	MeetingType       string
	MeetingTypeDesc   string
	Dates             string
	Days              string
	Time              string
	Location          string
	SiteCode          string
	PrimaryInstructor string
	Fee               string
	Comment           string
}

// ImportExcelSchedule imports course data from Excel file
func (scheduler *wmu_scheduler) ImportExcelSchedule(filePath string, schedule *Schedule) error {

	// Open the Excel file
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("error opening Excel file: %v", err)
	}
	defer f.Close()

	// Get the first sheet (CS)
	sheetName := f.GetSheetList()[0]

	// Get all rows
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("error reading Excel sheet: %v", err)
	}

	if len(rows) < 6 {
		return fmt.Errorf("insufficient data in Excel file")
	}

	// Headers are in row 5 (index 4)
	headers := rows[4]

	// Create a map of column indices
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[strings.TrimSpace(header)] = i
	}

	// Import courses starting from row 6 (index 5)
	var importedCount int
	var errorCount int

	for i := 5; i < len(rows); i++ {
		row := rows[i]

		// Skip empty rows
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" {
			continue
		}

		// Parse course data
		courseData := parseExcelRow(row, columnMap)

		// Skip rows that don't have CRN (likely comment rows)
		if courseData.CRN == "" || !isValidCRN(courseData.CRN) {
			continue
		}

		// Import the course
		err := scheduler.importCourseFromExcel(courseData, schedule)
		if err != nil {
			log.Printf("Error importing course CRN %s: %v", courseData.CRN, err)
			errorCount++
		} else {
			importedCount++
		}
	}

	log.Printf("Import completed: %d courses imported, %d errors", importedCount, errorCount)
	return nil
}

// parseExcelRow parses a row from Excel into ExcelCourseData
func parseExcelRow(row []string, columnMap map[string]int) ExcelCourseData {
	data := ExcelCourseData{}

	// Helper function to get value from column
	getValue := func(columnName string) string {
		if idx, exists := columnMap[columnName]; exists && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}

	data.CRN = getValue("CRN")
	data.CourseID = getValue("Course ID")
	data.Section = getValue("Section")
	data.Status = getValue("Status")
	data.Title = getValue("Title")
	data.Link1 = getValue("Link1")
	data.Link2 = getValue("Link2")
	data.SchedType = getValue("Sched Type")
	data.Reserved = getValue("Rsvrd")
	creditRange := getValue("Credit Hours")
	if strings.Contains(creditRange, "-") {
		parts := strings.Split(creditRange, "-")
		if len(parts) == 2 {
			data.MinCreditHours = strings.TrimSpace(parts[0])
			data.MaxCreditHours = strings.TrimSpace(parts[1])
		} else {
			data.MinCreditHours = creditRange
			data.MaxCreditHours = creditRange
		}
	} else {
		data.MinCreditHours = creditRange
		data.MaxCreditHours = creditRange
	}
	data.BillingHours = getValue("Billing Hours")
	contactRange := getValue("Contact Hours")
	if strings.Contains(contactRange, "-") {
		parts := strings.Split(contactRange, "-")
		if len(parts) == 2 {
			data.MinContactHours = strings.TrimSpace(parts[0])
			data.MaxContactHours = strings.TrimSpace(parts[1])
		} else {
			data.MinContactHours = contactRange
			data.MaxContactHours = contactRange
		}
	} else {
		data.MinContactHours = contactRange
		data.MaxContactHours = contactRange
	}
	data.Gradeable = getValue("Grad- able")
	data.Capacity = getValue("Cap")
	data.WaitlistCap = getValue("Waitlist Cap")
	data.SpecialApproval = getValue("Spec Appr")
	data.MeetingType = getValue("Mtg Type")
	data.MeetingTypeDesc = getValue("Meeting Type Desc")
	data.Dates = getValue("Dates")
	data.Days = getValue("Days")
	data.Time = getValue("Time")
	data.Location = getValue("Location")
	data.SiteCode = getValue("Site Code")
	data.PrimaryInstructor = getValue("Primary Instructor")
	data.Fee = getValue("Fee")
	data.Comment = getValue("Comment ")

	return data
}

// importCourseFromExcel imports a single course from Excel data
func (scheduler *wmu_scheduler) importCourseFromExcel(data ExcelCourseData, schedule *Schedule) error {
	// Parse course number and prefix from Course ID (e.g., "CS 1110")
	courseParts := strings.Fields(data.CourseID)
	if len(courseParts) < 2 {
		return fmt.Errorf("invalid course ID format: %s", data.CourseID)
	}
	// Check for duplicate schedule
	courseNum := 0
	courseNum, err := strconv.Atoi(courseParts[1])
	if err != nil {
		return fmt.Errorf("invalid course number in Course ID: %s", data.CourseID)
	}

	// Parse CRN
	crn, err := strconv.Atoi(data.CRN)
	if err != nil {
		return fmt.Errorf("invalid CRN: %s", data.CRN)
	}

	// Parse section
	section := data.Section

	// Parse credits
	minCredits, err := strconv.Atoi(data.MinCreditHours)
	if err != nil || minCredits < 0 {
		return fmt.Errorf("invalid credit hours: %s", data.MinCreditHours)
	}

	maxCredits, err := strconv.Atoi(data.MaxCreditHours)
	if err != nil || maxCredits < 0 {
		return fmt.Errorf("invalid credit hours: %s", data.MaxCreditHours)
	}

	// Parse contact hours
	minContactHours, err := strconv.Atoi(data.MinContactHours)
	if err != nil || minContactHours < 0 {
		return fmt.Errorf("invalid contact hours: %s", data.MinContactHours)
	}

	maxContactHours, err := strconv.Atoi(data.MaxContactHours)
	if err != nil || maxContactHours < 0 {
		return fmt.Errorf("invalid contact hours: %s", data.MaxContactHours)
	}

	// Parse capacity
	capacity, err := strconv.Atoi(data.Capacity)
	if err != nil || capacity < 0 {
		return fmt.Errorf("invalid capacity: %s", data.Capacity)
	}

	// Parse time slot
	timeSlotID := 0
	if data.Time != "" && data.Days != "" {
		timeSlotID, _ = scheduler.findOrCreateTimeSlot(data.Days, data.Time)
	}

	// Parse room
	roomID := 0
	if data.Location != "" {
		roomID, _ = scheduler.findOrCreateRoom(data.Location)
	}

	// Parse instructor
	instructorID := 0
	if data.PrimaryInstructor != "" {
		instructorID, _ = scheduler.findOrCreateInstructor(data.PrimaryInstructor)
	}

	// Parse section as int
	sectionInt, err := strconv.Atoi(section)
	if err != nil {
		return fmt.Errorf("invalid section: %s", section)
	}

	appr := 0
	if strings.TrimSpace(data.SpecialApproval) != "" {
		appr = 1
	}

	lab := 0
	if data.Link1 == "B1" && minCredits == 0 {
		lab = 1
	}

	err = scheduler.AddOrUpdateCourse(crn, sectionInt, schedule.ID, courseNum, data.Title,
		minCredits, maxCredits, minContactHours, maxContactHours, capacity, appr, lab, instructorID, roomID,
		timeSlotID, data.MeetingType, "Scheduled", data.Comment)

	return err
}

// Helper functions
func isValidCRN(crn string) bool {
	return len(crn) == 5 && isNumeric(crn)
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func parseTime(timeStr string) (string, error) {
	// Convert "1130" to "11:30:00"
	if len(timeStr) != 4 {
		return "", fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour := timeStr[:2]
	minute := timeStr[2:]
	return fmt.Sprintf("%s:%s:00", hour, minute), nil
}

// Web handler for Excel import
func (scheduler *wmu_scheduler) ImportExcelHandler(c *gin.Context) {
	// Handle file upload
	file, err := c.FormFile("excel_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Get form parameters
	term := c.PostForm("term")
	yearStr := c.PostForm("year")
	prefixName := c.PostForm("prefix")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	// Save uploaded file
	uploadPath := fmt.Sprintf("uploads/%s", file.Filename)
	err = c.SaveUploadedFile(file, uploadPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create schedule if it doesn't exist, otherwise get existing schedule
	schedule, err := scheduler.AddOrGetSchedule(term, year, prefixName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	// Import the Excel file
	err = scheduler.ImportExcelSchedule(uploadPath, schedule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Excel file imported successfully"})
}
