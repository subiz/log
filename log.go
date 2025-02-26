package log

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var w *Writer

func init() {
	hostname, _ := os.Hostname()
	service := strings.Split(hostname, "-")[0]
	w = &Writer{hostname: hostname, service: service}
}

// Err logs a message with severity LOG_ERR, ignoring the severity
// passed to New.
func Err(accid string, err error, v ...interface{}) error {
	// ignore if there is no error
	if err == nil {
		return nil
	}

	if accid == "" {
		accid = "subiz"
	}

	field := M{"account_id": accid}
	for i, vv := range v {
		key := strconv.Itoa(i)
		b, _ := json.Marshal(vv)
		field[key] = string(b)
	}

	// same as Error(err, field)
	// we copied this code to keep the stack unchanged
	outerr := NewError(err, field)
	b, _ := json.Marshal(outerr)
	w.writeAndRetry(accid, LOG_ERR, string(b))
	return outerr
}

func Info(v ...any) error {
	if len(v) == 0 {
		return nil
	}

	var accid string
	v0 := v[0]
	switch t := v0.(type) {
	case context.Context:
		ctx := t
		var msg = ""
		v = v[1:]
		if len(v) >= 1 {
			msg = fmt.Sprintf("%v", v[0])
			v = v[1:]
		}
		args := addStack(ctx, v)
		stdoutlog(ctx, slog.LevelInfo, msg, args...)
		logger.InfoContext(ctx, msg, args...)
		return nil
	case string:
		accid = t
		v = v[1:]
	default:
	}

	format := strings.Repeat("%v ", len(v))
	m := fmt.Sprintf(format, v...)
	_, err := w.writeAndRetry(accid, LOG_INFO, m)
	return err
}
