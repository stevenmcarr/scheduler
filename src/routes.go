package main

import (
	"github.com/gin-gonic/gin"
)

func (scheduler *wmu_scheduler) router() *gin.Engine {
	r := gin.Default()

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
