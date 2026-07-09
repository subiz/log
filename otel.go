package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var loggerProvider *sdklog.LoggerProvider
var traceProvider *sdktrace.TracerProvider
var logger *slog.Logger
var tracer trace.Tracer
var hostname string
var logServerHost string
var logServerSecret string

// otelDisabled reports whether OpenTelemetry tracing/log export should be
// turned off for this process. Set the standard OTEL_SDK_DISABLED=true to opt
// out — in that mode no spans are recorded/exported and error/log reports are
// delivered to /collect/ (the same lightweight pipeline used by log.Err)
// instead of the OTLP /v1/logs endpoint. Defaults to enabled so other services
// keep full OpenTelemetry unchanged.
func otelDisabled() bool {
	return strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true")
}

func init() {
	hostname, _ = os.Hostname()

	logServerHost = os.Getenv("LOG_SERVER_HOST")
	if logServerHost == "" {
		logServerHost = "https://log.subiz.net"
	}
	logServerSecret = os.Getenv("LOG_SERVER_SECRET")
	if logServerSecret != "" {
		go func() {
			for {
				time.Sleep(2 * time.Second)
				flushLog()
			}
		}()

		go func() {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			<-ctx.Done()
			flushLog()
			stop()
		}()
	}

	if otelDisabled() {
		// Lightweight mode: leave the global TracerProvider as a no-op so that
		// instrumentation (e.g. otelgrpc) creates non-recording spans and does
		// not allocate or export anything. Error/log reports go to /collect/
		// via collectHandler instead of the OTLP /v1/logs endpoint.
		tracer = noop.NewTracerProvider().Tracer("")
		otel.SetTracerProvider(noop.NewTracerProvider())
		logger = slog.New(&collectHandler{})
		return
	}

	loggerProvider = newLoggerProvider(nil)
	scope := ""
	logger = slog.New(otelslog.NewHandler(scope, otelslog.WithLoggerProvider(loggerProvider)))
	traceProvider = newTraceProvider()
	tracer = traceProvider.Tracer("")
	otel.SetTracerProvider(traceProvider)
}

var defaultsloghandler = slog.Default().Handler()

func Shutdown() {
	// Lightweight mode has no OTLP providers to flush; just drain /collect/.
	if loggerProvider != nil {
		if err := loggerProvider.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}
	}
	if traceProvider != nil {
		if err := traceProvider.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}
	}
	flushLog()
}

// collectHandler is an slog.Handler that queues records for delivery to the
// /collect/ endpoint instead of exporting them as OpenTelemetry logs. It keeps
// error/info reporting working without any OTLP/tracing dependency.
type collectHandler struct {
	attrs []slog.Attr
}

func (h *collectHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *collectHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	na := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	na = append(na, h.attrs...)
	na = append(na, attrs...)
	return &collectHandler{attrs: na}
}

func (h *collectHandler) WithGroup(_ string) slog.Handler { return h }

func newTraceProvider() *sdktrace.TracerProvider {
	opts := []sdktrace.TracerProviderOption{}

	// stdoutexporter, _ := stdouttrace.New()
	// opts = append(opts, sdktrace.WithBatcher(stdoutexporter))

	if logServerSecret != "" {
		httpexporter, err := otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpointURL(logServerHost+"/v1/traces"),
			otlptracehttp.WithHeaders(map[string]string{"secret": logServerSecret}))
		if err != nil {
			panic(err)
		}
		opts = append(opts, sdktrace.WithBatcher(httpexporter))
	}

	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("", attribute.String("host.name", hostname)),
	)
	if err != nil {
		panic(err)
	}

	opts = append(opts, sdktrace.WithResource(r))
	return sdktrace.NewTracerProvider(opts...)
}

func newLoggerProvider(res *resource.Resource) *sdklog.LoggerProvider {
	defres, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes("", attribute.String("host.name", hostname)))
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
		opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(httpexporter)))
	}

	// stdoutexporter := &slogExporter{}
	// opts = append(opts, log.WithProcessor(log.NewBatchProcessor(stdoutexporter)))
	return sdklog.NewLoggerProvider(opts...)
}

func (h *collectHandler) Handle(_ context.Context, r slog.Record) error {
	if logServerSecret == "" {
		return nil
	}

	level := "LOG"
	if r.Level >= slog.LevelError {
		level = "ERROR"
	}

	accid := "subiz"
	caller := "-"
	var msg strings.Builder
	msg.WriteString(r.Message)

	appendAttr := func(a slog.Attr) {
		switch a.Key {
		case "account_id":
			if s := a.Value.String(); s != "" {
				accid = s
			}
		case "_stack":
			// keep only the closest frame as the caller column; the full
			// stack is already embedded in error payloads.
			if s := a.Value.String(); s != "" {
				if i := strings.Index(s, " | "); i >= 0 {
					caller = s[:i]
				} else {
					caller = s
				}
			}
		default:
			msg.WriteByte(' ')
			msg.WriteString(a.Key)
			msg.WriteByte('=')
			msg.WriteString(a.Value.String())
		}
	}
	for _, a := range h.attrs {
		appendAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		appendAttr(a)
		return true
	})

	line := r.Time.Format("06-01-02 15:04:05") + " app " + level + " " + hostname + " " + accid + " " + caller + " " + msg.String()

	logmaplock.Lock()
	if len(logmap) < LIMIT_LOG_MAP_LENGTH {
		logmap = append(logmap, line)
	}
	logmaplock.Unlock()
	return nil
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	stdoutlog(ctx, slog.LevelDebug, msg, args...)
	logger.DebugContext(ctx, msg, args...)
}

// log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func stdoutlog(ctx context.Context, level slog.Level, msg string, args ...any) {
	var pc uintptr
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = defaultsloghandler.Handle(ctx, r)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	stdoutlog(ctx, slog.LevelError, msg, args...)
	logger.ErrorContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	stdoutlog(ctx, slog.LevelInfo, msg, args...)
	logger.InfoContext(ctx, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	// args = addStack(ctx, args)
	logger.LogAttrs(ctx, level, msg, attrs...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	args = addStack(ctx, args)
	logger.WarnContext(ctx, msg, args...)
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

// ctx, span := log.Start(ctx, "funcname")
// defer span.End()
func Start(ctx context.Context, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}
	_, spanName, _ := GetStack(-1)
	return tracer.Start(ctx, spanName, opts...)
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

	// force sample
	spanContext := trace.SpanContextFromContext(ctx)
	if ctx == nil {
		spanContext = trace.NewSpanContext(trace.SpanContextConfig{})
	}

	ctx = trace.ContextWithSpanContext(ctx,
		spanContext.WithTraceFlags(spanContext.TraceFlags().WithSampled(true)))
	ctx, span := tracer.Start(ctx, "track-"+code)
	defer span.End()

	wrapargs := []any{}

	wrapargs = append(wrapargs, "log_type", "track")

	wrapargs = append(wrapargs, args...)

	stack, funcname, _ := GetStack(-1)
	wrapargs = append(wrapargs, "_stack", stack)
	wrapargs = append(wrapargs, "function_name", funcname)
	wrapargs = append(wrapargs, "server_name", hostname)

	stdoutlog(ctx, slog.LevelInfo, "TRK-"+code, wrapargs...)
	logger.InfoContext(ctx, "TRK-"+code, wrapargs...)
}
