package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/go-programming-tour-book/blog-service/pkg/logger"
	"github.com/go-programming-tour-book/blog-service/pkg/tracing"
)

// newTestLogger returns a Logger that writes to buf so we can inspect JSON output.
func newTestLogger(buf *bytes.Buffer) *logger.Logger {
	return logger.NewLogger(buf, "", log.LstdFlags)
}

func TestJSONFormat_WithTraceID(t *testing.T) {
	traceID := tracing.NewTraceID()
	ctx := tracing.SetTraceID(context.Background(), traceID)

	var buf bytes.Buffer
	l := newTestLogger(&buf)
	l.WithContext(ctx).Infof("test message")

	var out map[string]interface{}
	if err := json.Unmarshal(extractJSON(t, buf.Bytes()), &out); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	got, ok := out["trace_id"]
	if !ok {
		t.Fatal("expected 'trace_id' field in log output, but it was absent")
	}
	if got != traceID {
		t.Errorf("trace_id = %q, want %q", got, traceID)
	}
}

func TestJSONFormat_WithoutContext(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)

	// No panic expected, no trace_id field expected.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Logger.Infof without context panicked: %v", r)
		}
	}()

	l.Infof("test message without context")

	var out map[string]interface{}
	if err := json.Unmarshal(extractJSON(t, buf.Bytes()), &out); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if _, ok := out["trace_id"]; ok {
		t.Error("did not expect 'trace_id' field when no context is set")
	}
}

func TestJSONFormat_EmptyTraceID(t *testing.T) {
	// Context is set but trace-id is absent — should not emit trace_id field.
	ctx := context.Background() // no trace-id set

	var buf bytes.Buffer
	l := newTestLogger(&buf)
	l.WithContext(ctx).Infof("test message empty trace")

	var out map[string]interface{}
	if err := json.Unmarshal(extractJSON(t, buf.Bytes()), &out); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}

	if _, ok := out["trace_id"]; ok {
		t.Error("did not expect 'trace_id' field when trace-id is empty")
	}
}

// extractJSON finds the JSON object in a log line (strips the stdlib log prefix).
func extractJSON(t *testing.T, raw []byte) []byte {
	t.Helper()
	start := bytes.IndexByte(raw, '{')
	if start == -1 {
		t.Fatalf("no JSON object found in log output: %q", raw)
	}
	return raw[start:]
}
