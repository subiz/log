package log_test

import (
	"testing"

	log "github.com/subiz/log"
)

// TestWrapNoDuplicateCode guards against a regression where wrapping an
// existing *AError appended the extra codes once per field, producing
// "internal,retryable,retryable,retryable" instead of "internal,retryable".
func TestWrapNoDuplicateCode(t *testing.T) {
	base := log.NewError(nil, nil, log.E_internal)
	wrapped := log.NewError(base, log.M{"a": "1", "b": "2", "c": "3"}, "retryable")

	if wrapped.Code != "internal,retryable" {
		t.Errorf("wrapped codes = %q, want internal,retryable", wrapped.Code)
	}
}

// TestMessageOverrideUnknownCode guards against a nil-map panic: overriding
// _message when the code has no ErrorTable entry used to write to a nil map.
func TestMessageOverrideUnknownCode(t *testing.T) {
	err := log.NewError(nil, log.M{"_message": map[string]string{"en_US": "custom"}}, "some_unknown_code")
	if err.Message["en_US"] != "custom" {
		t.Errorf("expected custom message, got %v", err.Message)
	}
}

func TestSpanName(t *testing.T) {
	_, spanName, _ := log.GetStack(-2)
	if spanName != "github.com/subiz/log_test.TestSpanName" {
		t.Errorf("caller name = %q, want github.com/subiz/log_test.TestSpanName", spanName)
	}
}
