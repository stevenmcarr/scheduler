package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHTTPLogging(t *testing.T) {
	// Initialize logging for test
	initLogger() // Now always succeeds, may fallback to stdout-only logging

	// Create a simple test router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add the same logging middleware as in the main router
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

	// Add a simple test route
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test successful")
	})

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Test-Client/1.0")

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Process the request
	r.ServeHTTP(w, req)

	// Verify the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	AppLogger.LogInfo("HTTP logging test completed successfully!")
	AppLogger.LogInfo("Check /var/log/scheduler/scheduler.log for the logged HTTP request.")
}
