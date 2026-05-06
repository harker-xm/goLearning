package tracing_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/go-programming-tour-book/blog-service/pkg/tracing"
)

// uuidV4Regex validates UUID v4 format.
var uuidV4Regex = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

func TestNewTraceID_Format(t *testing.T) {
	id := tracing.NewTraceID()
	if !uuidV4Regex.MatchString(id) {
		t.Errorf("NewTraceID() = %q, does not match UUID v4 format", id)
	}
}

func TestNewTraceID_Unique(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := tracing.NewTraceID()
		if _, dup := seen[id]; dup {
			t.Fatalf("NewTraceID() returned duplicate value %q at iteration %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestSetGetTraceID(t *testing.T) {
	want := "test-trace-id-123"
	ctx := tracing.SetTraceID(context.Background(), want)
	got := tracing.GetTraceID(ctx)
	if got != want {
		t.Errorf("GetTraceID() = %q, want %q", got, want)
	}
}

func TestGetTraceID_EmptyContext(t *testing.T) {
	got := tracing.GetTraceID(context.Background())
	if got != "" {
		t.Errorf("GetTraceID(emptyCtx) = %q, want empty string", got)
	}
}

func TestGetTraceID_NilContext(t *testing.T) {
	// GetTraceID must not panic on nil context.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetTraceID(nil) panicked: %v", r)
		}
	}()
	got := tracing.GetTraceID(nil)
	if got != "" {
		t.Errorf("GetTraceID(nil) = %q, want empty string", got)
	}
}
