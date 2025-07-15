package main

import (
	"html/template"
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
	r := gin.Default()

	// Define custom template functions
	r.SetFuncMap(template.FuncMap{
		"replace": func(old, new, s string) string {
			return strings.Replace(s, old, new, -1)
		},
	})

	// Load HTML templates - try multiple paths
	templatePaths := []string{
		"templates/*",     // when run from src directory
		"src/templates/*", // when run from root directory
		"*/templates/*",   // fallback pattern
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
		ErrorFunc: func(c *gin.Context) {
			c.JSON(400, gin.H{
				"error": "CSRF token mismatch",
				"debug": gin.H{
					"token_from_form":   c.PostForm("csrf_token"),
					"token_from_header": c.GetHeader("X-CSRF-Token"),
				},
			})
			c.Abort()
		},
	}))

	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// GET routes
	r.GET("/scheduler/signup", func(c *gin.Context) {
		scheduler.ShowSignupFormGin(c)
	})
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
	r.POST("/scheduler/signup", func(c *gin.Context) {
		scheduler.SignupUserGin(c)
	})

	// Navigation routes
	r.GET("/scheduler/courses", func(c *gin.Context) {
		scheduler.RenderCoursesPageGin(c)
	})
	r.GET("/scheduler/rooms", func(c *gin.Context) {
		scheduler.RenderRoomsPageGin(c)
	})
	r.GET("/scheduler/timeslots", func(c *gin.Context) {
		scheduler.RenderTimeslotsPageGin(c)
	})
	r.GET("/scheduler/instructors", func(c *gin.Context) {
		scheduler.RenderInstructorsPageGin(c)
	})
	r.GET("/scheduler/departments", func(c *gin.Context) {
		scheduler.RenderDepartmentsPageGin(c)
	})
	r.GET("/scheduler/prefixes", func(c *gin.Context) {
		scheduler.RenderPrefixesPageGin(c)
	})
	r.GET("/scheduler/users", func(c *gin.Context) {
		scheduler.RenderUsersPageGin(c)
	})

	r.GET("/scheduler/delete", func(c *gin.Context) {
		scheduler.DeleteScheduleGin(c)
	})
	// Logout route
	r.GET("/scheduler/logout", func(c *gin.Context) {
		// Clear session cookie
		c.SetCookie("session", "", -1, "/", "", false, true)
		// Redirect to login page
		c.Redirect(http.StatusFound, "/scheduler/login")
	})

	return r
}
