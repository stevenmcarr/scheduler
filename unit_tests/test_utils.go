package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var routeCounter int64

// Helper function to create authenticated session for testing
func CreateAuthenticatedSession(router *gin.Engine, username string) *http.Cookie {
	// Use atomic counter to ensure unique route names
	counter := atomic.AddInt64(&routeCounter, 1)
	routeName := fmt.Sprintf("/test-auth-%d", counter)

	loginHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("username", username)
		session.Save()
		c.JSON(http.StatusOK, gin.H{"status": "logged in"})
	}

	router.POST(routeName, loginHandler)

	req, _ := http.NewRequest("POST", routeName, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	if len(cookies) > 0 {
		return cookies[0]
	}
	return nil
}

// Helper function to setup test router with sessions
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Note: CSRF middleware disabled for testing to simplify test setup
	// In production, CSRF protection should always be enabled

	return router
}
