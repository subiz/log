package log

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

var hostname string
var w *Writer

var service string

func init() {
	hostname, _ := os.Hostname()
	service = strings.Split(hostname, "-")[0]
	fmt.Println("CONFIG LOGING FOR SERVICE", service)
	dial("", "", LOG_WARNING|LOG_DAEMON)
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

	tag := service + "." + accid
	_, err = w.writeAndRetry(tag, LOG_ERR, m)
	return err
}

func Info(accid string, v ...interface{}) error {
	format := strings.Repeat("%v ", len(v))
	m := fmt.Sprintf(format, v...)
	// extract accid from context
	tag := service + "." + accid
	_, err := w.writeAndRetry(tag, LOG_INFO, m)
	return err
}

// Dial establishes a connection to a log daemon by connecting to
// address raddr on the specified network. Each write to the returned
// writer sends a log message with the facility and severity
// (from priority) and tag. If tag is empty, the os.Args[0] is used.
// If network is empty, Dial will connect to the local syslog server.
// Otherwise, see the documentation for net.Dial for valid values
// of network and raddr.
func dial(network, raddr string, priority Priority) (*Writer, error) {
	if priority < 0 || priority > LOG_LOCAL7|LOG_DEBUG {
		return nil, errors.New("log/syslog: invalid priority")
	}

	// set global variable
	w = &Writer{
		priority: priority,
		hostname: hostname,
		network:  network,
		raddr:    raddr,
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.connect()
	if err != nil {
		return nil, err
	}
	return w, err
}

// unixSyslog opens a connection to the syslog daemon running on the
// local machine using a Unix domain socket.

func unixSyslog() (conn serverConn, err error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
	for _, network := range logTypes {
		for _, path := range logPaths {
			conn, err := net.Dial(network, path)
			if err == nil {
				return &netConn{conn: conn, local: true}, nil
			}
		}
	}
	return nil, errors.New("Unix syslog delivery error")
}
