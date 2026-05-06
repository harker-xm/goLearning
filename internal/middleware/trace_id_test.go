package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-programming-tour-book/blog-service/internal/middleware"
	"github.com/go-programming-tour-book/blog-service/pkg/tracing"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(middleware.TraceID())
	r.GET("/test", handler)
	return r
}

func TestTraceIDMiddleware_GeneratesNewID(t *testing.T) {
	var capturedID string
	r := newTestRouter(func(c *gin.Context) {
		capturedID = tracing.GetTraceID(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if capturedID == "" {
		t.Error("expected a generated trace-id in context, got empty string")
	}
	if w.Header().Get("X-Trace-ID") == "" {
		t.Error("expected X-Trace-ID response header to be set")
	}
}

func TestTraceIDMiddleware_UsesExistingHeader(t *testing.T) {
	const existingID = "my-existing-trace-id"
	var capturedID string
	r := newTestRouter(func(c *gin.Context) {
		capturedID = tracing.GetTraceID(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", existingID)
	r.ServeHTTP(w, req)

	if capturedID != existingID {
		t.Errorf("context trace-id = %q, want %q", capturedID, existingID)
	}
}

func TestTraceIDMiddleware_SetsResponseHeader(t *testing.T) {
	const existingID = "resp-header-test-id"
	r := newTestRouter(func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", existingID)
	r.ServeHTTP(w, req)

	got := w.Header().Get("X-Trace-ID")
	if got != existingID {
		t.Errorf("X-Trace-ID response header = %q, want %q", got, existingID)
	}
}

func TestTraceIDMiddleware_InjectsContext(t *testing.T) {
	const wantID = "ctx-inject-test-id"
	var gotFromCtx string
	r := newTestRouter(func(c *gin.Context) {
		gotFromCtx = tracing.GetTraceID(c.Request.Context())
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Trace-ID", wantID)
	r.ServeHTTP(w, req)

	if gotFromCtx != wantID {
		t.Errorf("tracing.GetTraceID(ctx) = %q, want %q", gotFromCtx, wantID)
	}
}
