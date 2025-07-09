package main

import (
	"github.com/gin-gonic/gin"
)

func (scheduler *wmu_scheduler) router() *gin.Engine {
	r := gin.Default()
	r.POST("/scheduler/signup", func(c *gin.Context) {
		scheduler.ShowSignupForm(c.Writer, c.Request)
	})
	r.GET("/scheduler/login", func(c *gin.Context) {
		scheduler.ShowLoginForm(c.Writer, c.Request)
	})
	r.GET("/scheduler", func(c *gin.Context) {
		loggedIn, err := scheduler.CheckSession(c.Writer, c.Request)
		if err != nil {
			c.String(500, "Error checking session: %v", err)
			return
		}
		if !loggedIn {
			c.Redirect(302, "/scheduler/login")
			return
		}
		c.String(200, "Welcome to the scheduler!")
	})
	return r
}
