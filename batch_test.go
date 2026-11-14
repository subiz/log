package log

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func TestPartialBatchExportsBeforeShutdown(t *testing.T) {
	// Explicit library settings must win over SDK environment defaults.
	t.Setenv("OTEL_BLRP_SCHEDULE_DELAY", "100")
	exported := make(chan time.Time, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/logs" {
			t.Errorf("unexpected export path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		exported <- time.Now()
	}))
	defer server.Close()
	oldHost, oldSecret := logServerHost, logServerSecret
	logServerHost, logServerSecret = server.URL, "local-test"
	t.Cleanup(func() { logServerHost, logServerSecret = oldHost, oldSecret })
	started := time.Now()
	provider := newLoggerProvider(nil)
	defer provider.Shutdown(context.Background())
	l := slog.New(otelslog.NewHandler("test", otelslog.WithLoggerProvider(provider)))
	l.Info("partial batch")
	select {
	case at := <-exported:
		if elapsed := at.Sub(started); elapsed < 1500*time.Millisecond {
			t.Fatalf("exported too early: %s; expected 2-second interval", elapsed)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("partial batch was not exported before shutdown")
	}
}

func TestShutdownContextDeadline(t *testing.T) {
	// Simulate a collector that accepts the request but never responds in time.
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release
	}))
	defer server.Close()
	defer close(release)
	oldHost, oldSecret, oldProvider := logServerHost, logServerSecret, loggerProvider
	logServerHost, logServerSecret = server.URL, "local-test"
	t.Cleanup(func() { logServerHost, logServerSecret, loggerProvider = oldHost, oldSecret, oldProvider })
	loggerProvider = newLoggerProvider(nil)
	l := slog.New(otelslog.NewHandler("test", otelslog.WithLoggerProvider(loggerProvider)))
	l.Info("pending log")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- ShutdownContext(ctx) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected caller's deadline error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ShutdownContext ignored the caller's deadline")
	}
}
