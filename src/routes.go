package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func (scheduler *wmu_scheduler) router() *gin.Engine {
	// Create a new Gin engine without default middleware
	r := gin.New()

	// Configure Gin to use our custom logger
	logFile, err := os.OpenFile("/var/log/scheduler/scheduler.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		AppLogger.LogError("Failed to open log file for Gin", err)
		// Fall back to stdout only
		gin.DefaultWriter = os.Stdout
	} else {
		// Set Gin to write to both stdout and log file
		gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	}

	// Add custom logging middleware with detailed request information using AppLogger
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			// Use AppLogger for HTTP logging instead of default Gin logger
			AppLogger.LogHTTP(
				param.Method,
				param.Path,
				param.ClientIP,
				param.Request.UserAgent(),
				param.StatusCode,
				param.Latency.String(),
			)
			return "" // Return empty string since we're using AppLogger directly
		},
		Output: gin.DefaultWriter, // Still set an output for any fallback logging
	}))

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Define custom template functions
	r.SetFuncMap(template.FuncMap{
		"replace": func(old, new, s string) string {
			return strings.Replace(s, old, new, -1)
		},
	})

	// Load HTML templates - try multiple paths
	templatePaths := []string{
		"src/templates/*", // when run from root directory
	}

	var templatePattern string
	for _, pattern := range templatePaths {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			templatePattern = pattern
			break
		}
	}

	if templatePattern == "" {
		// If no pattern works, try to find the directory
		if _, err := os.Stat("templates"); err == nil {
			templatePattern = "templates/*"
		} else if _, err := os.Stat("src/templates"); err == nil {
			templatePattern = "src/templates/*"
		} else {
			panic("Could not find templates directory")
		}
	}

	r.LoadHTMLGlob(templatePattern)

	// Serve static files (images, CSS, JS, etc.)
	r.Static("/scheduler/images", "./images")

	// Add session middleware (required for CSRF)
	// Use a 32-byte key for AES-256 encryption
	sessionKey := []byte("your-secret-key-for-sessions-32b")
	store := cookie.NewStore(sessionKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})
	r.Use(sessions.Sessions("session", store))

	// Add CSRF middleware
	r.Use(csrf.Middleware(csrf.Options{
		Secret: "b7f8c2e4a1d9f3e6c5b2a8d7e4f1c3b6", // generated 32-char random hex string
		TokenGetter: func(c *gin.Context) string {
			// Look for CSRF token in form field 'csrf_token' or header 'X-CSRF-Token'
			token := c.PostForm("csrf_token")
			if token == "" {
				token = c.GetHeader("X-CSRF-Token")
			}
			return token
		},
		ErrorFunc: func(c *gin.Context) {
			// For login errors, redirect back to form with error message
			if c.Request.URL.Path == "/scheduler/login" {
				c.Redirect(http.StatusSeeOther, "/scheduler/login?error=Invalid+request")
				c.Abort()
				return
			}
			// For other routes, return JSON error
			c.JSON(400, gin.H{
				"error": "CSRF token mismatch",
				"debug": gin.H{
					"token_from_form":   c.PostForm("csrf_token"),
					"token_from_header": c.GetHeader("X-CSRF-Token"),
					"expected_token":    csrf.GetToken(c),
					"path":              c.Request.URL.Path,
					"method":            c.Request.Method,
				},
			})
			c.Abort()
		},
	}))

	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// GET routes
	r.GET("/scheduler/login", func(c *gin.Context) {
		scheduler.ShowLoginFormGin(c)
	})
	r.GET("/scheduler", func(c *gin.Context) {
		scheduler.RenderHomePageGin(c)
	})
	r.GET("/scheduler/test", func(c *gin.Context) {
		schedules, err := scheduler.GetAllSchedules()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading schedules: "+err.Error())
			return
		}

		data := gin.H{
			"Schedules": schedules,
			"User":      &User{Username: "testuser", Email: "test@example.com", Administrator: true},
		}

		c.HTML(http.StatusOK, "home.html", data)
	})

	// Excel import routes
	r.GET("/scheduler/import", func(c *gin.Context) {
		scheduler.ShowImportPage(c)
	})
	r.POST("/scheduler/import", func(c *gin.Context) {
		scheduler.ImportExcelHandler(c)
	})

	// POST routes
	r.POST("/scheduler/login", func(c *gin.Context) {
		scheduler.LoginUserGin(c)
	})

	// Navigation routes
	r.GET("/scheduler/courses", func(c *gin.Context) {
		scheduler.RenderCoursesPageGin(c)
	})
	r.POST("/scheduler/courses", func(c *gin.Context) {
		scheduler.SaveCoursesGin(c)
	})

	r.GET("/scheduler/add_course", func(c *gin.Context) {
		scheduler.RenderAddCoursePageGin(c)
	})

	r.POST("/scheduler/add_course", func(c *gin.Context) {
		scheduler.AddCourseGin(c)
	})

	// Export courses to Excel
	r.GET("/scheduler/export/:scheduleID", func(c *gin.Context) {
		scheduler.ExportCoursesToExcel(c)
	})

	// Course schedule table view
	r.GET("/scheduler/courses_table", func(c *gin.Context) {
		scheduler.RenderCoursesTableGin(c)
	})

	r.GET("/scheduler/rooms", func(c *gin.Context) {
		scheduler.RenderRoomsPageGin(c)
	})

	r.POST("/scheduler/rooms", func(c *gin.Context) {
		scheduler.SaveOrDeleteRoomsGin(c)
	})

	r.GET("/scheduler/add_room", func(c *gin.Context) {
		scheduler.RenderAddRoomPageGin(c)
	})

	r.POST("/scheduler/add_room", func(c *gin.Context) {
		scheduler.AddRoomGin(c)
	})

	r.GET("/scheduler/timeslots", func(c *gin.Context) {
		scheduler.RenderTimeslotsPageGin(c)
	})

	r.POST("/scheduler/timeslots", func(c *gin.Context) {
		scheduler.SaveTimeslotsGin(c)
	})

	r.GET("/scheduler/add_timeslot", func(c *gin.Context) {
		scheduler.RenderAddTimeslotPageGin(c)
	})
	r.POST("/scheduler/add_timeslot", func(c *gin.Context) {
		scheduler.AddTimeslotGin(c)
	})

	r.GET("/scheduler/instructors", func(c *gin.Context) {
		scheduler.RenderInstructorsPageGin(c)
	})
	r.POST("/scheduler/instructors", func(c *gin.Context) {
		scheduler.SaveInstructorsGin(c)
	})
	r.GET("/scheduler/departments", func(c *gin.Context) {
		scheduler.RenderDepartmentsPageGin(c)
	})
	r.POST("/scheduler/departments", func(c *gin.Context) {
		scheduler.SaveDepartmentsGin(c)
	})
	r.GET("/scheduler/add_department", func(c *gin.Context) {
		scheduler.RenderAddDepartmentPageGin(c)
	})
	r.POST("/scheduler/add_department", func(c *gin.Context) {
		scheduler.AddDepartmentGin(c)
	})
	r.GET("/scheduler/prefixes", func(c *gin.Context) {
		scheduler.RenderPrefixesPageGin(c)
	})
	r.POST("/scheduler/prefixes", func(c *gin.Context) {
		scheduler.SavePrefixesGin(c)
	})
	r.GET("/scheduler/add_prefix", func(c *gin.Context) {
		scheduler.RenderAddPrefixPageGin(c)
	})
	r.POST("/scheduler/add_prefix", func(c *gin.Context) {
		scheduler.AddPrefixGin(c)
	})
	r.GET("/scheduler/users", func(c *gin.Context) {
		scheduler.RenderUsersPageGin(c)
	})
	r.POST("/scheduler/users", func(c *gin.Context) {
		scheduler.SaveUsersGin(c)
	})

	r.GET("/scheduler/add_user", func(c *gin.Context) {
		scheduler.RenderAddUserPageGin(c)
	})

	r.POST("/scheduler/add_user", func(c *gin.Context) {
		scheduler.AddUserGin(c)
	})

	r.POST("/scheduler/delete_schedule", func(c *gin.Context) {
		scheduler.DeleteScheduleGin(c)
	})
	// Logout route
	r.GET("/scheduler/logout", func(c *gin.Context) {
		scheduler.LogoutUserGin(c)
	})

	r.GET("/scheduler/add_instructor", func(c *gin.Context) {
		scheduler.RenderAddInstructorPageGin(c)
	})
	r.POST("/scheduler/add_instructor", func(c *gin.Context) {
		scheduler.AddInstructorGin(c)
	})

	// Session message routes
	r.POST("/scheduler/set_error_message", func(c *gin.Context) {
		scheduler.SetErrorMessageGin(c)
	})

	return r
}
