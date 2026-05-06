package middleware

import "github.com/gin-gonic/gin"

// AppInfo injects application name and version into the request context.
func AppInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("app_name", "blog-service")
		c.Set("app_version", "1.0.0")
		c.Next()
	}
}
