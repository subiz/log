package log_test

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/subiz/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	collector "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/proto"
)

// A fresh process exercises package init with a configured log destination,
// independently of other tests calling Shutdown on the global logger.
func TestLogsWithoutTraceExport(t *testing.T) {
	if os.Getenv("SUBIZ_LOG_TEST_CHILD") == "1" {
		runLogsWithoutTraces(t)
		return
	}
	for _, mode := range []struct {
		name, sampler, disabled, path string
	}{
		{name: "default", path: "/v1/logs"},
		{name: "always_on", sampler: "always_on", path: "/v1/logs"},
		{name: "collect", sampler: "always_on", disabled: "true", path: "/collect/"},
	} {
		t.Run(mode.name, func(t *testing.T) {
			var mu sync.Mutex
			paths := map[string]int{}
			var messages []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				paths[r.URL.Path]++
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				switch r.URL.Path {
				case "/v1/logs":
					var request collector.ExportLogsServiceRequest
					if err := proto.Unmarshal(body, &request); err != nil {
						t.Error(err)
						w.WriteHeader(400)
						return
					}
					for _, resource := range request.ResourceLogs {
						for _, scope := range resource.ScopeLogs {
							for _, record := range scope.LogRecords {
								messages = append(messages, record.Body.GetStringValue())
							}
						}
					}
					w.Header().Set("Content-Type", "application/x-protobuf")
				case "/collect/":
					for _, line := range strings.Fields(string(body)) {
						decoded, err := base64.StdEncoding.DecodeString(line)
						if err != nil {
							t.Error(err)
							w.WriteHeader(400)
							return
						}
						messages = append(messages, string(decoded))
					}
				default:
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestLogsWithoutTraceExport$")
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "LOG_SERVER_") && !strings.HasPrefix(entry, "OTEL_") && !strings.HasPrefix(entry, "SUBIZ_LOG_TEST_CHILD=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "SUBIZ_LOG_TEST_CHILD=1", "LOG_SERVER_HOST="+server.URL,
				"LOG_SERVER_SECRET=local-test", "OTEL_TRACES_SAMPLER="+mode.sampler, "OTEL_SDK_DISABLED="+mode.disabled)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("child failed: %v\n%s", err, output)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(paths) != 1 || paths[mode.path] == 0 {
				t.Fatalf("unexpected export destinations: %v", paths)
			}
			joined := strings.Join(messages, "\n")
			for _, want := range []string{"ordinary-log-probe", "TRK-track-probe", "error-probe"} {
				if !strings.Contains(joined, want) {
					t.Errorf("missing %q in exported logs: %s", want, joined)
				}
			}
		})
	}
}

func runLogsWithoutTraces(t *testing.T) {
	defer log.Shutdown()
	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{1}, SpanID: trace.SpanID{1}, TraceFlags: trace.FlagsSampled, Remote: true,
	})
	for _, ctx := range []context.Context{nil, context.Background(), trace.ContextWithRemoteSpanContext(context.Background(), parent)} {
		ctx, span := log.Start(ctx)
		if span.IsRecording() {
			t.Error("log.Start records spans")
		}
		log.SetAttributes(span, "key", "value")
		log.SetSpanAttributes(ctx, "key", "value")
		_, globalSpan := otel.Tracer("test").Start(ctx, "instrumentation")
		if globalSpan.IsRecording() {
			t.Error("global instrumentation records spans")
		}
		globalSpan.End()
		span.End()
	}
	log.InfoContext(context.Background(), "ordinary-log-probe")
	log.Track(nil, "track-probe")
	log.Error(context.Background(), nil, "internal", "probe", "error-probe")
}
