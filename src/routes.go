package main

import (
	"github.com/gin-gonic/gin"
)

func (scheduler *wmu_scheduler) router() *gin.Engine {
	r := gin.Default()

	// Apply Apache2 proxy middleware to all routes
	r.Use(scheduler.Apache2ProxyMiddleware())
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
		c.String(200, "Welcome to the scheduler!")
	})
	return r
}
