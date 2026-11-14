//go:build !windows && !plan9

package log_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestImportPreservesSignalExit(t *testing.T) {
	signals := map[string]syscall.Signal{"interrupt": syscall.SIGINT, "terminate": syscall.SIGTERM}
	if name := os.Getenv("SUBIZ_LOG_SIGNAL_TEST"); name != "" {
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		if err := process.Signal(signals[name]); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)
		t.Fatal("the logging package swallowed the termination signal")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for name, signal := range signals {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestImportPreservesSignalExit$")
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "LOG_") && !strings.HasPrefix(entry, "OTEL_") && !strings.HasPrefix(entry, "SUBIZ_LOG_") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			// No logs are emitted, so this configured exporter never sends a request.
			cmd.Env = append(cmd.Env, "LOG_SERVER_SECRET=local-test", "LOG_SERVER_HOST=http://127.0.0.1:1", "SUBIZ_LOG_SIGNAL_TEST="+name)
			output, err := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("expected signal exit, got %v: %s", err, output)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != signal {
				t.Fatalf("expected exit from %s, got %v: %s", signal, exit, output)
			}
		})
	}
}
