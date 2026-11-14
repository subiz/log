package log_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	log "github.com/subiz/log"
	"go.opentelemetry.io/otel/trace"
)

func TestWrappedError(t *testing.T) {
	root := errors.New("connection refused")
	base := log.NewError(root, log.M{"id": "original", "_detail": "private"}, log.E_invalid_input)
	wrapped := fmt.Errorf("request failed: %w", base)
	if !log.IsErr(wrapped, "invalid_input") || log.IsErr(wrapped, "internal") {
		t.Fatal("IsErr does not follow the wrapped error's codes")
	}
	for _, wrap := range []struct {
		name string
		fn   func(error) *log.AError
	}{
		{"NewError", func(err error) *log.AError {
			return log.NewError(err, log.M{"attempt": 2}, log.E_internal, log.E_retryable, log.E_internal)
		}},
		{"NewError2", func(err error) *log.AError {
			return log.NewError2(context.Background(), err, []log.E{log.E_internal, log.E_retryable, log.E_internal}, "attempt", 2)
		}},
	} {
		t.Run(wrap.name, func(t *testing.T) {
			got := wrap.fn(wrapped)
			if got.Code != "invalid_input,internal,retryable" || got.Class != 500 {
				t.Fatalf("wrong classification: code=%q class=%d", got.Code, got.Class)
			}
			if got.Id != base.Id || got.Number != strings.Replace(base.Number, "SBQ-", "SBR-", 1) {
				t.Fatalf("lost originating identity: id=%q number=%q", got.Id, got.Number)
			}
			if got.Attrs["id"].Value != "original" || got.Attrs["attempt"] == nil || got.XHiddenAttrs["detail"].Value != "private" {
				t.Fatal("lost error attributes while wrapping")
			}
			if !errors.Is(got, root) || !errors.Is(got, base) || errors.Unwrap(got) != wrapped {
				t.Fatal("lost the in-process cause chain")
			}
			if !errors.Is(got.Clone(), root) {
				t.Fatal("Clone lost the cause chain")
			}
			if base.Code != "invalid_input" || base.Class != 400 || base.Attrs["attempt"] != nil || !strings.HasPrefix(base.Number, "SBQ-") {
				t.Fatal("wrapping mutated the base error")
			}
		})
	}
}

func TestErrorSerialization(t *testing.T) {
	root := errors.New("private connection details")
	base := log.NewError(root, log.M{"id": "record-1", "_secret": "private-token"}, log.E_internal)
	public := base.ToJSON()
	if !json.Valid([]byte(public)) || strings.Contains(public, "private") {
		t.Fatalf("public JSON is invalid or exposes internal data: %s", public)
	}
	var payload struct {
		Class int        `json:"class"`
		Error log.AError `json:"error"`
	}
	if err := json.Unmarshal([]byte(public), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Class != 500 || payload.Error.Code != "internal" || payload.Error.Attrs["id"].Value != "record-1" {
		t.Fatalf("missing public error metadata: %s", public)
	}
	// Existing serialized errors must still wrap even without an in-process cause.
	decoded := log.NewError(errors.New(base.Error()), nil, log.E_retryable)
	if decoded.Id != base.Id || decoded.Code != "internal,retryable" || decoded.XHiddenAttrs["root"].Value != root.Error() {
		t.Fatal("lost metadata when wrapping a serialized error")
	}
	if !json.Valid([]byte(strings.TrimPrefix(decoded.Error(), "#ERR "))) {
		t.Fatal("the in-process cause must not change the JSON representation")
	}
}

func TestNewError2PreservesOriginalTrace(t *testing.T) {
	span := trace.NewSpanContext(trace.SpanContextConfig{TraceID: trace.TraceID{1}, SpanID: trace.SpanID{2}})
	ctx := trace.ContextWithSpanContext(context.Background(), span)
	base := log.NewError2(ctx, nil, []log.E{log.E_internal})
	wrapped := log.NewError2(context.Background(), fmt.Errorf("outer: %w", base), []log.E{log.E_retryable})
	if wrapped.TraceId != span.TraceID().String() || wrapped.SpanId != span.SpanID().String() {
		t.Fatal("wrapping replaced the original trace")
	}
}

func TestIsErrNil(t *testing.T) {
	var typedNil *log.AError
	for _, err := range []error{nil, typedNil, errors.New("ordinary error")} {
		if log.IsErr(err, "internal") {
			t.Fatalf("unexpected match for %v", err)
		}
	}
}
