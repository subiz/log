package log_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

// A fresh process exercises package init with each destination and environment.
func TestLogsWithoutTraceExport(t *testing.T) {
	if os.Getenv("SUBIZ_LOG_TEST_CHILD") == "1" {
		runLogsWithoutTraces(t)
		return
	}
	for _, mode := range []struct {
		name, sampler, disabled, path, hostname string
	}{
		{name: "default", path: "/v1/logs"},
		{name: "always_on", sampler: "always_on", path: "/v1/logs"},
		{name: "pod_hostname", hostname: "zaloperson-75fb677958-65dz6", path: "/v1/logs"},
		{name: "replica_hostname", hostname: "convo-1", path: "/v1/logs"},
		{name: "sdk_disabled", sampler: "always_on", disabled: "true", path: "/v1/logs"},
	} {
		t.Run(mode.name, func(t *testing.T) {
			var mu sync.Mutex
			paths := map[string]int{}
			var messages []string
			var batches []int
			var legacyCount int
			wantHost := mode.hostname
			if wantHost == "" {
				var err error
				wantHost, err = os.Hostname()
				if err != nil {
					t.Fatal(err)
				}
			}
			wantService := strings.Split(wantHost, "-")[0]
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
					batchSize := 0
					for _, resource := range request.ResourceLogs {
						attrs := map[string]string{}
						for _, attr := range resource.GetResource().GetAttributes() {
							attrs[attr.Key] = attr.Value.GetStringValue()
						}
						if attrs["host.name"] != wantHost || attrs["service.name"] != wantService {
							t.Errorf("resource identity: host=%q service=%q; want host=%q service=%q", attrs["host.name"], attrs["service.name"], wantHost, wantService)
						}
						for _, scope := range resource.ScopeLogs {
							for _, record := range scope.LogRecords {
								messages = append(messages, record.Body.GetStringValue())
								batchSize++
								attrs := map[string]string{}
								for _, attr := range record.Attributes {
									attrs[attr.Key] = attr.Value.GetStringValue()
								}
								body := record.Body.GetStringValue()
								if strings.Contains(body, "legacy-") {
									legacyCount++
									if attrs["account_id"] != "test-account" || attrs["_stack"] == "" || attrs["server_name"] != wantHost {
										t.Errorf("missing legacy metadata: %v", attrs)
									}
									if strings.Contains(body, "legacy-error") && record.SeverityNumber < 17 {
										t.Error("lost error severity")
									}
								}
								if body == "TRK-track-probe" && (attrs["log_type"] != "track" || attrs["_stack"] == "") {
									t.Error("lost tracking metadata")
								}
								if body == "plain-error-probe" && record.SeverityNumber != 17 {
									t.Error("lost plain error severity")
								}

							}
						}
					}
					batches = append(batches, batchSize)
					w.Header().Set("Content-Type", "application/x-protobuf")
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
				if !strings.HasPrefix(entry, "LOG_SERVER_") && !strings.HasPrefix(entry, "LOG_HOSTNAME=") && !strings.HasPrefix(entry, "OTEL_") && !strings.HasPrefix(entry, "SUBIZ_LOG_TEST_CHILD=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "SUBIZ_LOG_TEST_CHILD=1", "LOG_SERVER_HOST="+server.URL,
				"LOG_SERVER_SECRET=local-test", "LOG_HOSTNAME="+mode.hostname, "OTEL_TRACES_SAMPLER="+mode.sampler, "OTEL_SDK_DISABLED="+mode.disabled)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("child failed: %v\n%s", err, output)
			}
			mu.Lock()
			defer mu.Unlock()
			if len(paths) != 1 || paths[mode.path] == 0 {
				t.Fatalf("unexpected export destinations: %v", paths)
			}
			if legacyCount != 3 {
				t.Errorf("legacy logs=%d, want 3", legacyCount)
			}
			if len(messages) != 211 {
				t.Errorf("records=%d, want 211 (no losses or duplicates)", len(messages))
			}
			var sawFullBatch bool
			for _, size := range batches {
				if size > 100 {
					t.Errorf("oversized batch: %d", size)
				}
				if size == 100 {
					sawFullBatch = true
				}
			}
			if !sawFullBatch {
				t.Errorf("no full batch: %v", batches)
			}
			joined := strings.Join(messages, "\n")
			for _, want := range []string{"ordinary-log-probe", "TRK-track-probe", "error-probe", "legacy-info", "legacy-error", "plain-error-probe", "warn-probe", "attrs-probe"} {
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
	log.Info("test-account", "legacy-info")
	log.Err("test-account", errors.New("legacy-error"))
	log.Error3("test-account", errors.New("legacy-error3"), nil)
	log.Info(context.Background(), "context-info-probe")
	log.ErrorContext(context.Background(), "plain-error-probe")
	log.WarnContext(context.Background(), "warn-probe")
	log.LogAttrs(context.Background(), slog.LevelError, "attrs-probe", slog.String("key", "value"))
	for i := 0; i < 201; i++ {
		log.InfoContext(context.Background(), fmt.Sprintf("batch-probe-%d", i))
	}
	log.InfoContext(context.Background(), "ordinary-log-probe")
	log.Track(nil, "track-probe")
	log.Error(context.Background(), nil, "internal", "probe", "error-probe")
}
