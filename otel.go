package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var loggerProvider *sdklog.LoggerProvider
var logger *slog.Logger
var tracer = noop.NewTracerProvider().Tracer("")
var logServerHost string
var logServerSecret string

const maxQueuedLogRecords = 100_000

func init() {
	logServerHost = os.Getenv("LOG_SERVER_HOST")
	if logServerHost == "" {
		logServerHost = "https://log.subiz.net"
	}
	logServerSecret = os.Getenv("LOG_SERVER_SECRET")
	// The log server only accepts logs. Keep existing instrumentation safe to
	// call without recording spans or creating a trace exporter.
	otel.SetTracerProvider(noop.NewTracerProvider())

	loggerProvider = newLoggerProvider(nil)
	scope := ""
	logger = slog.New(otelslog.NewHandler(scope, otelslog.WithLoggerProvider(loggerProvider)))
}

// Shutdown drains pending logs with a 10-second deadline. Applications own
// signal handling; use ShutdownContext to choose a deadline and handle errors.
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ShutdownContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// ShutdownContext drains the shared OTLP queue and stops exporting logs.
// Stop log producers before calling it. A canceled context may leave logs unsent;
// subsequent calls do not retry shutdown. A nil context uses context.Background.
func ShutdownContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if loggerProvider == nil {
		return nil
	}
	return loggerProvider.Shutdown(ctx)
}

func newLoggerProvider(res *resource.Resource) *sdklog.LoggerProvider {
	defres, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes("",
			attribute.String("host.name", hostname),
			attribute.String("service.name", w.service)))
	if err != nil {
		panic(err)
	}

	opts := []sdklog.LoggerProviderOption{sdklog.WithResource(defres)}
	if res != nil {
		opts = append(opts, sdklog.WithResource(res))
	}
	if logServerSecret != "" {
		httpexporter, err := otlploghttp.New(context.Background(),
			otlploghttp.WithEndpointURL(logServerHost+"/v1/logs"),
			otlploghttp.WithHeaders(map[string]string{"secret": logServerSecret}))
		if err != nil {
			panic(err)
		}
		opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(httpexporter,
			sdklog.WithExportMaxBatchSize(100),
			sdklog.WithExportInterval(2*time.Second),
			sdklog.WithMaxQueueSize(maxQueuedLogRecords),
		)))
	}

	return sdklog.NewLoggerProvider(opts...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	emit(ctx, slog.LevelDebug, msg, args...)
}

// emit sends the same message and attributes to the console and OTLP.
func emit(ctx context.Context, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	stdoutlog(ctx, level, msg, args...)
	logger.Log(ctx, level, msg, args...)
}

func stdoutlog(ctx context.Context, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	handler := slog.Default().Handler()
	if !handler.Enabled(ctx, level) {
		return
	}
	r := slog.NewRecord(time.Now(), level, msg, 0)
	r.Add(args...)
	_ = handler.Handle(ctx, r)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	emit(ctx, slog.LevelError, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	emit(ctx, slog.LevelInfo, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	args := make([]any, 0, len(attrs)+2)
	for _, attr := range attrs {
		args = append(args, attr)
	}
	args = addStack(ctx, args)
	emit(ctx, level, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	emit(ctx, slog.LevelWarn, msg, args...)
}

func SetAttributes(span trace.Span, attrs ...any) {
	for i := 0; i < len(attrs)-1; i += 2 {
		keyi := attrs[i]
		if keyi == nil {
			continue
		}
		key, ok := keyi.(string)
		if !ok {
			continue
		}
		val := attrs[i+1]
		attr := convertValue(key, val)
		span.SetAttributes(attr)
	}
}

func SetSpanAttributes(ctx context.Context, attrs ...any) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	for i := 0; i < len(attrs)-1; i += 2 {
		keyi := attrs[i]
		if keyi == nil {
			continue
		}
		key, ok := keyi.(string)
		if !ok {
			continue
		}
		val := attrs[i+1]
		attr := convertValue(key, val)
		span.SetAttributes(attr)
	}
}

// func  With(args ...any) *Logger {}
// func  WithGroup(name string) *Logger {}

// Start returns a non-recording span for compatibility with existing callers.
// It does not create or export a distributed trace.
func Start(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	return tracer.Start(ctx, "", opts...)
}

func addStack(ctx context.Context, args []any) []any {
	stack, _, _ := GetStack(0)
	args = append(args, "_stack")
	args = append(args, stack)
	return args
}

func convertValue(k string, v any) attribute.KeyValue {
	// Handling the most common types without reflect is a small perf win.
	switch val := v.(type) {
	case bool:
		return attribute.Bool(k, val)
	case string:
		return attribute.String(k, val)
	case int:
		return attribute.Int(k, val)
	case int8:
		return attribute.Int(k, int(val))
	case int16:
		return attribute.Int(k, int(val))
	case int32:
		return attribute.Int(k, int(val))
	case int64:
		return attribute.Int64(k, val)
	case uint:
		return attribute.Int64(k, int64(val))
	case uint8:
		return attribute.Int64(k, int64(val))
	case uint16:
		return attribute.Int64(k, int64(val))
	case uint32:
		return attribute.Int64(k, int64(val))
	case uint64:
		return attribute.Int64(k, int64(val))
	case float32:
		return attribute.Float64(k, float64(val))
	case float64:
		return attribute.Float64(k, float64(val))
	case time.Duration:
		return attribute.Int64(k, val.Nanoseconds())
	case time.Time:
		return attribute.Int64(k, val.UnixNano())
	case []int:
		return attribute.IntSlice(k, val)
	case []int64:
		return attribute.Int64Slice(k, val)
	case []string:
		return attribute.StringSlice(k, val)
	case []float64:
		return attribute.Float64Slice(k, val)
	case []bool:
		return attribute.BoolSlice(k, val)
	case error:
		return attribute.String(k, val.Error())
	}
	return attribute.String(k, fmt.Sprintf("%v", v))
}

// log.Track(nil, "dupplicated_email", "title", "duplicated email")
// log.Track(nil, "dupplicated_email", "title", "duplicated email", "noti", false})
func Track(ctx context.Context, code string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}

	wrapargs := []any{}

	wrapargs = append(wrapargs, "log_type", "track")

	wrapargs = append(wrapargs, args...)

	stack, funcname, _ := GetStack(-1)
	wrapargs = append(wrapargs, "_stack", stack)
	wrapargs = append(wrapargs, "function_name", funcname)
	wrapargs = append(wrapargs, "server_name", hostname)

	emit(ctx, slog.LevelInfo, "TRK-"+code, wrapargs...)
}

// K create index key
func K(key string) string {
	return "key=" + key
}
