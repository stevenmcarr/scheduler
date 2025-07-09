package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Apache2ProxyMiddleware ensures only Apache2 can act as a proxy
func (scheduler *wmu_scheduler) Apache2ProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Apache2 specific headers
		userAgent := c.GetHeader("User-Agent")
		server := c.GetHeader("Server")

		// Check if request is coming through Apache2 proxy
		isApache2Proxy := strings.Contains(userAgent, "Apache") ||
			strings.Contains(server, "Apache") ||
			c.GetHeader("X-Forwarded-Server") != "" // Apache mod_proxy sets this

		// Allow direct access only if it's from localhost (for development)
		remoteAddr := c.ClientIP()
		isDirect := remoteAddr == "127.0.0.1" || remoteAddr == "::1" || remoteAddr == "localhost"

		if !isApache2Proxy && !isDirect {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Only Apache2 proxy is allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecureProxyMiddleware validates a custom header set by Apache2
func (scheduler *wmu_scheduler) SecureProxyMiddleware(secretToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for custom header that Apache2 should set
		proxyToken := c.GetHeader("X-Proxy-Token")

		// Allow direct access only if it's from localhost (for development)
		remoteAddr := c.ClientIP()
		isDirect := remoteAddr == "127.0.0.1" || remoteAddr == "::1" || remoteAddr == "localhost"

		if proxyToken != secretToken && !isDirect {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Invalid proxy token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Alternative: IP-based restriction
func (scheduler *wmu_scheduler) IPBasedProxyMiddleware(allowedProxyIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if request is from allowed proxy IPs
		isAllowed := false
		for _, allowedIP := range allowedProxyIPs {
			if clientIP == allowedIP {
				isAllowed = true
				break
			}
		}

		// Allow localhost for development
		if clientIP == "127.0.0.1" || clientIP == "::1" {
			isAllowed = true
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: Unauthorized proxy IP",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
