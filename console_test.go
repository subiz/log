package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestConsoleOutput(t *testing.T) {
	oldDefault, oldLogger := slog.Default(), logger
	t.Cleanup(func() { slog.SetDefault(oldDefault); logger = oldLogger })
	// Exercise console output independently of exporter configuration.
	logger = slog.New(slog.DiscardHandler)
	ctx := context.Background()
	for _, tc := range []struct {
		name, level string
		emit        func()
	}{
		{"debug", "DEBUG", func() { DebugContext(ctx, "probe", "key", "value") }},
		{"info", "INFO", func() { InfoContext(ctx, "probe", "key", "value") }},
		{"info-overload", "INFO", func() { Info(ctx, "probe", "key", "value") }},
		{"warn", "WARN", func() { WarnContext(ctx, "probe", "key", "value") }},
		{"error", "ERROR", func() { ErrorContext(ctx, "probe", "key", "value") }},
		{"attrs", "ERROR", func() { LogAttrs(ctx, slog.LevelError, "probe", slog.String("key", "value")) }},
		{"track", "INFO", func() { Track(nil, "probe", "key", "value") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			slog.SetDefault(slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug})))
			tc.emit()
			var record map[string]any
			if err := json.Unmarshal(out.Bytes(), &record); err != nil {
				t.Fatalf("expected one JSON record, got %q: %v", out.String(), err)
			}
			if record["level"] != tc.level || record["key"] != "value" || record["_stack"] == nil || record["_stack"] == "" {
				t.Fatalf("incomplete record: %v", record)
			}
		})
	}

	t.Run("error-cause", func(t *testing.T) {
		var out bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&out, nil)))
		Error(nil, errors.New("database connection refused"), "internal", "account_id", "ac123")
		var record struct {
			Message string `json:"msg"`
		}
		if err := json.Unmarshal(out.Bytes(), &record); err != nil {
			t.Fatal(err)
		}
		var payload AError
		if err := json.Unmarshal([]byte(record.Message), &payload); err != nil {
			t.Fatalf("console must include the structured error: %s", out.String())
		}
		if payload.Code != "internal" || payload.XHiddenAttrs["root"] == nil || payload.XHiddenAttrs["root"].Value != "database connection refused" || payload.XHiddenAttrs["stack"] == nil {
			t.Fatalf("console lost error details: %s", out.String())
		}
	})

	t.Run("configured-level", func(t *testing.T) {
		var out bytes.Buffer
		slog.SetDefault(slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelWarn})))
		DebugContext(ctx, "filtered")
		InfoContext(ctx, "filtered")
		if out.Len() != 0 {
			t.Fatalf("ignored the console handler's level: %s", out.String())
		}
	})
}

func TestInfoStackLimit(t *testing.T) {
	oldDefault, oldLogger := slog.Default(), logger
	t.Cleanup(func() { slog.SetDefault(oldDefault); logger = oldLogger })
	slog.SetDefault(slog.New(slog.DiscardHandler))
	ctx := context.Background()
	for _, tc := range []struct {
		name    string
		limited bool
		emit    func()
	}{
		{"legacy-info", true, func() { Info("test-account", "probe") }},
		{"info", true, func() { Info(ctx, "probe") }},
		{"info-context", true, func() { InfoContext(ctx, "probe") }},
		{"info-attrs", true, func() { LogAttrs(ctx, slog.LevelInfo, "probe") }},
		{"track", false, func() { Track(ctx, "probe") }},
		{"error-context", false, func() { ErrorContext(ctx, "probe") }},
		{"error-attrs", false, func() { LogAttrs(ctx, slog.LevelError, "probe") }},
		{"error", false, func() { Error(ctx, errors.New("probe"), "internal") }},
		{"legacy-error", false, func() { Err("test-account", errors.New("probe")) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			logger = slog.New(slog.NewJSONHandler(&out, nil))
			emitWithDeepStack(5, tc.emit)
			var record map[string]any
			if err := json.Unmarshal(out.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			stack, _ := record["_stack"].(string)
			if stack == "" {
				var payload AError
				if err := json.Unmarshal([]byte(record["msg"].(string)), &payload); err != nil {
					t.Fatal(err)
				}
				stack = payload.XHiddenAttrs["stack"].Value
			}
			frames := strings.Split(stack, " | ")
			if tc.limited && len(frames) != 3 {
				t.Fatalf("got %d frames, want 3: %s", len(frames), stack)
			}
			if !tc.limited && len(frames) <= 3 {
				t.Fatalf("full stack was truncated: %s", stack)
			}
			if !strings.Contains(frames[0], "console_test.go:") {
				t.Fatalf("stack lost the closest caller: %s", stack)
			}
		})
	}
}

//go:noinline
func emitWithDeepStack(depth int, emit func()) {
	if depth == 0 {
		emit()
		return
	}
	emitWithDeepStack(depth-1, emit)
}
