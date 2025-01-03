// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !windows && !plan9
// +build !windows,!plan9

package log

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// The Priority is a combination of the syslog facility and
// severity. For example, LOG_ALERT | LOG_FTP sends an alert severity
// message from the FTP facility. The default severity is LOG_EMERG;
// the default facility is LOG_KERN.
type Priority int

const severityMask = 0x07
const facilityMask = 0xf8

const (
	// Severity.

	// From /usr/include/sys/syslog.h.
	// These are the same on Linux, BSD, and OS X.
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

const (
	// Facility.

	// From /usr/include/sys/syslog.h.
	// These are the same up to LOG_FTP on Linux, BSD, and OS X.
	LOG_KERN Priority = iota << 3
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_ // unused
	_ // unused
	_ // unused
	_ // unused
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

// A Writer is a connection to a syslog server.
type Writer struct {
	hostname string
	service  string
	network  string
	raddr    string

	mu   sync.Mutex // guards conn
	conn net.Conn
}

// connect makes a connection to the syslog server.
// It must be called with w.mu held.
func (w *Writer) connect() (err error) {
	if w.conn != nil {
		// ignore err from close, it makes sense to continue anyway
		w.conn.Close()
		w.conn = nil
	}

	if w.network == "" {
		w.conn, err = unixSyslog()
		if w.hostname == "" {
			w.hostname = "localhost"
		}
	} else {
		var c net.Conn
		c, err = net.Dial(w.network, w.raddr)
		if err == nil {
			w.conn = c
			if w.hostname == "" {
				w.hostname = c.LocalAddr().String()
			}
		}
	}
	return
}

// Close closes a connection to the syslog daemon.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	return nil
}

func (w *Writer) writeAndRetry(accid string, p Priority, msg string) (int, error) {
	if accid == "" {
		accid = "subiz"
	}
	thep := facilityMask | (p & severityMask)

	w.mu.Lock()
	defer w.mu.Unlock()

	// ensure it ends in a \n
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}

	// log to stdout first
	caller := getCaller()
	now := time.Now()
	timestamp := now.Format(time.Stamp)
	fmt.Fprintf(os.Stdout, "<%d>%s %s %s[%s]: %s| %s%s",
		thep, timestamp, w.hostname, accid, w.service, caller, msg, nl)

	if logServerSecret != "" {
		level := "LOG"
		if p == LOG_ERR {
			level = "ERROR"
		}

		line := now.Format("06-01-02 15:04:05") + " app " + level + " " + w.hostname + " " + accid + " " + caller + " " + msg
		logmaplock.Lock()
		if len(logmap) < LIMIT_LOG_MAP_LENGTH {
			logmap = append(logmap, line)
		}
		logmaplock.Unlock()
	}

	if w.conn == nil {
		return 0, nil
	}

	// write generates and writes a syslog formatted string. The
	// format is as follows: <PRI>TIMESTAMP HOSTNAME TAG[PID]: MSG
	return fmt.Fprintf(w.conn, "<%d>%s %s[%s]: %s| %s%s",
		thep, timestamp, accid, w.service, caller, msg, nl)
}

// unixSyslog opens a connection to the syslog daemon running on the
// local machine using a Unix domain socket.

func unixSyslog() (conn net.Conn, err error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
	for _, network := range logTypes {
		for _, path := range logPaths {
			conn, err := net.Dial(network, path)
			if err == nil {
				return conn, nil
			}
		}
	}
	return nil, errors.New("unix syslog delivery error")
}
