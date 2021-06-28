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

	// establishes a connection to a log daemon by connecting to
	// address raddr on the specified network. Each write to the returned
	// writer sends a log message with the facility and severity
	// (from priority) and tag. If tag is empty, the os.Args[0] is used.
	// If network is empty, Dial will connect to the local syslog server.
	// Otherwise, see the documentation for net.Dial for valid values
	// of network and raddr.

	// set global variable
	w = &Writer{priority: LOG_WARNING | LOG_DAEMON, hostname: hostname, service: service}

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.connect(); err != nil {
		fmt.Println("ERR CANNOT CONNECT TO LOGGING SERVICE", err)
	}
}

// Err logs a message with severity LOG_ERR, ignoring the severity
// passed to New.
func Err(accid string, err error, v ...interface{}) error {
	// ignore if there is no error
	if err == nil {
		return nil
	}

	format := strings.Repeat("%v ", len(v))
	format = "ERR %s [stack %s] " + format
	v = append([]interface{}{err.Error(), getStack(2)}, v...)
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
