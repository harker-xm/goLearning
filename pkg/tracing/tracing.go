// Package tracing provides lightweight trace-id utilities for context propagation.
package tracing

import (
	"context"
	"crypto/rand"
	"fmt"
)

// contextKey is an unexported type to avoid collision with keys from other packages.
type contextKey struct{}

// NewTraceID generates a new random UUID v4 string.
func NewTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	// Set version bits to 4 (UUID v4).
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant bits to 10xx (RFC 4122).
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:],
	)
}

// SetTraceID stores traceID in ctx and returns the derived context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKey{}, traceID)
}

// GetTraceID retrieves the trace-id from ctx.
// Returns an empty string when ctx is nil or the value is absent.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(contextKey{}).(string)
	return v
}
