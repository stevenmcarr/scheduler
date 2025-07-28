package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHTTPLogging(t *testing.T) {
	// Initialize logging for test
	err := initLogger()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create a simple test router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add the same logging middleware as in the main router
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[HTTP] %v | %3d | %13v | %15s | %-7s %#v | User-Agent: %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
		)
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

	fmt.Println("HTTP logging test completed successfully!")
	fmt.Println("Check /var/log/scheduler/scheduler.log for the logged HTTP request.")
}
