package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	_ "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var loggerProvider *log.LoggerProvider
var traceProvider *sdktrace.TracerProvider
var logger *slog.Logger
var tracer trace.Tracer

var hostname string
var logServerHost string
var logServerSecret string

func init() {
	logServerHost = os.Getenv("LOG_SERVER_HOST")
	if logServerHost == "" {
		logServerHost = "https://log.subiz.net"
	}
	logServerSecret = os.Getenv("LOG_SERVER_SECRET")
	hostname, _ = os.Hostname()
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

	loggerProvider = newLoggerProvider(nil)
	logger = newLogger("")
	traceProvider = newTraceProvider()
	tracer = traceProvider.Tracer("")
}

var defaultsloghandler = slog.Default().Handler()

func newLogger(scope string) *slog.Logger {
	return slog.New(otelslog.NewHandler(scope, otelslog.WithLoggerProvider(loggerProvider)))
}

func Shutdown() {
	if err := loggerProvider.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}

	if err := traceProvider.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func Init(scope string, res *resource.Resource) {
	loggerProvider = newLoggerProvider(res)
	// Register as global logger provider so that it can be accessed global.LoggerProvider.
	// Most log bridges use the global logger provider as default.
	// If the global logger provider is not set then a no-op implementation
	// is used, which fails to generate data.
	global.SetLoggerProvider(loggerProvider)

	// Create an *slog.Logger and use it in your application.
	logger = newLogger(scope)
	traceProvider = newTraceProvider()
	otel.SetTracerProvider(traceProvider)
	tracer = traceProvider.Tracer(scope)
}

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
	hostname, _ := os.Hostname()
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.HostName(hostname),
		),
	)
	if err != nil {
		panic(err)
	}

	opts = append(opts, sdktrace.WithResource(r))
	return sdktrace.NewTracerProvider(opts...)
}

func newLoggerProvider(res *resource.Resource) *log.LoggerProvider {
	hostname, _ := os.Hostname()
	defres, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.HostName(hostname),
		))
	if err != nil {
		panic(err)
	}

	opts := []log.LoggerProviderOption{log.WithResource(defres)}
	if res != nil {
		opts = append(opts, log.WithResource(res))
	}
	if logServerSecret != "" {
		httpexporter, err := otlploghttp.New(context.Background(),
			otlploghttp.WithEndpointURL(logServerHost+"/v1/logs"),
			otlploghttp.WithHeaders(map[string]string{"secret": logServerSecret}))
		if err != nil {
			panic(err)
		}
		opts = append(opts, log.WithProcessor(log.NewBatchProcessor(httpexporter)))
	}

	// stdoutexporter := &slogExporter{}
	// opts = append(opts, log.WithProcessor(log.NewBatchProcessor(stdoutexporter)))
	return log.NewLoggerProvider(opts...)
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

func SetAttributesContext(ctx context.Context, attrs ...any) {
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

// trace
func Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, spanName, opts...)
}
func addStack(ctx context.Context, args []any) []any {
	stack, _ := GetStack(0)
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

	wrapargs = append(wrapargs, "log_type")
	wrapargs = append(wrapargs, "track")

	wrapargs = append(wrapargs, args...)

	wrapargs = addStack(ctx, wrapargs)

	stdoutlog(ctx, slog.LevelInfo, "TRK-"+code, wrapargs...)
	logger.InfoContext(ctx, "TRK-"+code, wrapargs...)
}
