package log

import (
	"fmt"
	"os"
	"strings"
)

var w *Writer

func init() {
	hostname, _ := os.Hostname()
	service := strings.Split(hostname, "-")[0]
	w = &Writer{priority: LOG_WARNING | LOG_DAEMON, hostname: hostname, service: service}
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

	format := strings.Repeat("%v ", len(v))
	format = "ERR %s [stack %s] " + format
	stack, _ := getStack(1)
	v = append([]interface{}{err.Error(), stack}, v...)
	m := fmt.Sprintf(format, v...)

	_, err = w.writeAndRetry(accid, LOG_ERR, m)
	return err
}

func Info(accid string, v ...interface{}) error {
	format := strings.Repeat("%v ", len(v))
	m := fmt.Sprintf(format, v...)
	_, err := w.writeAndRetry(accid, LOG_INFO, m)
	return err
}
