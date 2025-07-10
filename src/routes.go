package main

import (
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

func (scheduler *wmu_scheduler) router() *gin.Engine {
	r := gin.Default()

	// Add CSRF middleware
	r.Use(csrf.Middleware(csrf.Options{
		Secret: "b7f8c2e4a1d9f3e6c5b2a8d7e4f1c3b6", // generated 32-char random hex string
		ErrorFunc: func(c *gin.Context) {
			c.String(400, "CSRF token mismatch")
			c.Abort()
		},
	}))

	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// GET routes
	r.GET("/scheduler/signup", func(c *gin.Context) {
		scheduler.ShowSignupForm(c.Writer, c.Request)
	})
	r.GET("/scheduler/login", func(c *gin.Context) {
		scheduler.ShowLoginForm(c.Writer, c.Request)
	})
	r.GET("/scheduler", func(c *gin.Context) {
		loggedIn := scheduler.CheckSession(c.Writer, c.Request)
		if !loggedIn {
			c.Redirect(302, "/scheduler/login")
			return
		}
		scheduler.RenderHomePage(c.Writer, c.Request)
	})

	// POST routes
	r.POST("/scheduler/login", func(c *gin.Context) {
		scheduler.LoginUser(c.Writer, c.Request)
	})
	r.POST("/scheduler/signup", func(c *gin.Context) {
		scheduler.SignupUser(c.Writer, c.Request)
	})

	return r
}
