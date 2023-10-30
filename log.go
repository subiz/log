package log

import (
	"encoding/json"
	"fmt"
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
	m := fmt.Sprintf("%s %s", "error_"+outerr.Number, string(b))
	w.writeAndRetry(accid, LOG_ERR, m)
	return outerr
}

func Info(accid string, v ...interface{}) error {
	format := strings.Repeat("%v ", len(v))
	m := fmt.Sprintf(format, v...)
	_, err := w.writeAndRetry(accid, LOG_INFO, m)
	return err
}
