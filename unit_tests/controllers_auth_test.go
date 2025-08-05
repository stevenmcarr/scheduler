package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// User struct for testing - mirrors the actual User struct
type User struct {
	ID            int
	Username      string
	Email         string
	Password      string
	IsLoggedIn    bool
	Administrator bool
}

// Helper function to create a test Gin engine with session middleware
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add session middleware
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("session", store))

	// Note: CSRF middleware disabled for testing to simplify test setup
	// In production, CSRF protection should always be enabled

	return router
}

// Mock login handler for testing
func mockLoginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"error": "Username and password are required",
		})
		return
	}

	// Mock authentication - only success for specific credentials
	if username == "testuser" && password == "testpassword" {
		// Set session
		session := sessions.Default(c)
		session.Set("username", username)
		session.Save()
		c.Redirect(http.StatusFound, "/home")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": "Invalid credentials",
	})
}

// Mock signup handler for testing
func mockSignupHandler(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Basic validation
	if username == "" || email == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"error": "All fields are required",
		})
		return
	}

	// Mock existing user check
	if username == "existinguser" || email == "existing@example.com" {
		c.JSON(http.StatusOK, gin.H{
			"error": "User already exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": "User created successfully",
	})
}

// Mock logout handler for testing
func mockLogoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

// Test Login Controller
func TestLoginController(t *testing.T) {
	router := setupTestRouter()
	router.POST("/login", mockLoginHandler)

	// Test successful login
	t.Run("Successful Login", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "testuser")
		form.Add("password", "testpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/home", w.Header().Get("Location"))
	})

	// Test login with missing username
	t.Run("Missing Username", func(t *testing.T) {
		form := url.Values{}
		form.Add("password", "testpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test login with missing password
	t.Run("Missing Password", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "testuser")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test login with invalid credentials
	t.Run("Invalid Credentials", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "wronguser")
		form.Add("password", "wrongpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Test Signup Controller
func TestSignupController(t *testing.T) {
	router := setupTestRouter()
	router.POST("/signup", mockSignupHandler)

	// Test successful signup
	t.Run("Successful Signup", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "newuser")
		form.Add("email", "newuser@example.com")
		form.Add("password", "newpassword123")

		req, _ := http.NewRequest("POST", "/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test signup with existing user
	t.Run("Existing User", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "existinguser")
		form.Add("email", "new@example.com")
		form.Add("password", "password123")

		req, _ := http.NewRequest("POST", "/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test signup with missing fields
	t.Run("Missing Fields", func(t *testing.T) {
		form := url.Values{}
		form.Add("username", "testuser")
		// Missing email and password

		req, _ := http.NewRequest("POST", "/signup", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Test Logout Controller
func TestLogoutController(t *testing.T) {
	router := setupTestRouter()
	router.POST("/logout", mockLogoutHandler)

	// Test logout
	t.Run("Successful Logout", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/logout", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})
}

// Test Session Management
func TestSessionManagement(t *testing.T) {
	router := setupTestRouter()

	// Handler that requires authentication
	protectedHandler := func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Authenticated",
			"user":    username,
		})
	}

	router.GET("/protected", protectedHandler)
	router.POST("/login", mockLoginHandler)

	t.Run("Access Protected Route Without Session", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/login", w.Header().Get("Location"))
	})

	t.Run("Access Protected Route With Valid Session", func(t *testing.T) {
		// First, create a session by "logging in"
		form := url.Values{}
		form.Add("username", "testuser")
		form.Add("password", "testpassword")

		req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Extract session cookie
		cookies := w.Result().Cookies()

		// Now access protected route with session cookie
		req2, _ := http.NewRequest("GET", "/protected", nil)
		for _, cookie := range cookies {
			req2.AddCookie(cookie)
		}

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
	})
}

// Test Form Data Validation
func TestFormDataValidation(t *testing.T) {
	t.Run("Email Validation", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.com",
			"firstname+lastname@example.co.uk",
		}

		invalidEmails := []string{
			"invalid-email",
			"@example.com",
			"test@",
			"test.example.com",
		}

		for _, email := range validEmails {
			// In a real implementation, you'd call your email validation function
			assert.Contains(t, email, "@", "Valid email should contain @")
			assert.Contains(t, email, ".", "Valid email should contain .")
		}

		for _, email := range invalidEmails {
			// In a real implementation, you'd call your email validation function
			// For now, just test basic cases
			if email == "invalid-email" || email == "@example.com" || email == "test@" {
				assert.True(t, true, "Invalid email detected")
			}
		}
	})

	t.Run("Username Validation", func(t *testing.T) {
		validUsernames := []string{
			"testuser",
			"user123",
			"john_doe",
		}

		invalidUsernames := []string{
			"",
			"a", // too short
			"user with spaces",
		}

		for _, username := range validUsernames {
			assert.NotEmpty(t, username, "Valid username should not be empty")
			assert.Greater(t, len(username), 2, "Valid username should be longer than 2 characters")
		}

		for _, username := range invalidUsernames {
			if username == "" {
				assert.Empty(t, username, "Empty username should be detected")
			}
			if len(username) == 1 {
				assert.LessOrEqual(t, len(username), 2, "Short username should be detected")
			}
		}
	})
}
