package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-programming-tour-book/blog-service/pkg/tracing"
)

// TraceID returns a Gin middleware that propagates trace-ids through the request lifecycle.
//
// Behaviour:
//   - If the incoming request carries an "X-Trace-ID" header, that value is reused.
//   - Otherwise a new UUID v4 is generated via tracing.NewTraceID.
//   - The trace-id is stored in c.Request.Context() so all downstream code can call
//     tracing.GetTraceID(ctx) to retrieve it.
//   - The trace-id is also written to the "X-Trace-ID" response header so callers can
//     correlate logs with the request.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = tracing.NewTraceID()
		}

		ctx := tracing.SetTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}
