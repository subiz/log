package log_test

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/subiz/log"
	"go.opentelemetry.io/otel/trace"
)

func A() error {
	err := B()
	return err
}

func B() error {
	err := CCCCCC()
	return err
}

func CCCCCC() error {
	err := DDDDDD()
	return err
}

func DDDDDD() error {
	err := E()
	return err
}

func E() error {
	err := log.EServer(nil, log.M{"number3": "thanh"})
	return err
}

func TestError(t *testing.T) {
	// err := log.EInvalidZaloToken("thanh", "3290323", "Dayladau") // A()
	err := A()
	fmt.Println("EEEEEE", err.Error())
}

func TestLogErr(t *testing.T) {
	err := log.EInvalidGoogleToken("thanh", "3290323", "Dayladau") // A()
	log.Err("subiz", err, "param")
}

func TestWrap(t *testing.T) {
	var err error = log.EAccountLocked("thanh") // A()
	log.WrapStack(err, 0)
	time.Sleep(20 * time.Second)
}

func TestToJson(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()
	str := err.ToJSON()
	fmt.Println("JSON:", str)
}

func TestMarshalJson(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()
	str := log.NewError(err, nil).ToJSON()
	fmt.Println("JSON--------:", str)
}

func TestDontModify(t *testing.T) {
	err := log.EAccountLocked("thanh") // A()

	log.EServer(err, log.M{"secret": "2340asdffEIFJ42"})
	str := log.NewError(err, nil).ToJSON()
	if strings.Contains(str, "2340asdffEIFJ42") {
		t.Error("should not contain")
	}
}

func A2(ctx context.Context) error {
	err := B2(ctx)
	return err
}

func B2(ctx context.Context) error {
	err := CCCCCC2(ctx)
	return err
}

func CCCCCC2(ctx context.Context) error {
	err := DDDDDD2(ctx)
	return err
}

func DDDDDD2(ctx context.Context) error {
	err := E2(ctx)
	return err
}

func E2(ctx context.Context) error {
	err := log.Error(ctx, nil, "internal", "number3", "thanh")
	return err
}

func TestNoSpanError(t *testing.T) {
	defer log.Shutdown()
	// no trace must work

	// err := log.EInvalidZaloToken("thanh", "3290323", "Dayladau") // A()
	ctx := context.Background()
	ctx = context.WithValue(ctx, "account_id", "abcsble")
	err := A2(ctx)
	fmt.Println("EEEEEE", err.Error())
}

func TestSpanError(t *testing.T) {
	defer log.Shutdown()
	ctx, span := log.Start(context.Background())
	defer span.End()

	// err := log.EInvalidZaloToken("thanh", "3290323", "Dayladau") // A()
	ctx = context.WithValue(ctx, "account_id", "abcsble")
	err := A2(ctx)
	fmt.Println("EEEEEE", err.Error())
}

func TestSpanInfo(t *testing.T) {
	defer log.Shutdown()
	ctx, span := log.Start(context.Background())
	defer span.End()

	// err := log.EInvalidZaloToken("thanh", "3290323", "Dayladau") // A()
	log.SetSpanAttributes(ctx, "account_id", "abcsble")
	log.InfoContext(ctx, "xin chao the gioi")
	log.Info(ctx, "xin chao ca nuoc")
}

func parent(ctx context.Context) {
	ctx, span := log.Start(ctx)
	defer span.End()

	log.SetAttributes(span, "isTrue", true, "stringAttr", "hi!")
	// same as
	// span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "hi!"))

	child(ctx)
}

func child(ctx context.Context) {
	ctx, span := log.Start(ctx)
	defer span.End()

	log.InfoContext(ctx, "hello world", "user-id", "string", "user-count", 1995)
}

func TestTrace(t *testing.T) {
	defer log.Shutdown()

	ctx, span := log.Start(context.Background())
	defer span.End()

	parent(ctx)
}

func TestTrack(t *testing.T) {
	defer log.Shutdown()
	ctx, span := log.Start(context.Background())
	defer span.End()
	log.Track(ctx, "dup-email", "account_id", "sble4")
}

func TestSpanName(t *testing.T) {
	_, spanName, _ := log.GetStack(-2)
	if spanName != "github.com/subiz/log_test.TestSpanName" {
		t.Errorf("SHOULDEQ, GOT %s", spanName)
	}
}

func TestLoopLog(t *testing.T) {
	defer log.Shutdown()
	i := 0
	for {
		if i%100 == 99 {
			time.Sleep(10 * time.Second)
		}
		i++
		time.Sleep(10 * time.Millisecond)
		ctx, span := log.Start(context.Background())
		spanCtx := trace.SpanContextFromContext(ctx)

		fmt.Println("SSSS", spanCtx.TraceID().String())
		// if spanCtx.HasTraceID() {
		// traceID := spanCtx.TraceID()
		// return traceID.String()
		//}

		log.Track(ctx, "test", "account_id", "sble4", "i", "["+strconv.Itoa(i)+"]", "tag", "llm")
		span.End()
	}
}

func TestGenErrCode(t *testing.T) {
	// 3692131565 10
	// D48E40F6   16
	// 3E12T7D    32
	// 1P276Y5    34
	fmt.Println(strings.ToUpper(strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte("Invoke/service_unavailable,internal,retryable"))), 10)))
}
