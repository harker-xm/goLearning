package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// ContextTimeout sets a unified timeout on the request context.
// Downstream handlers should respect ctx.Done() / ctx.Err() to honour the deadline.
func ContextTimeout(t time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), t)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
