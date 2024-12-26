package log

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

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
}

var logmaplock = &sync.Mutex{}
var logmap = []string{}

// 1 line => 500 charater => 100_000line =>50 mb
const LIMIT_LOG_MAP_LENGTH = 100_000

func flushLog() {
	// flush periodically in 5s
	for {
		logmaplock.Lock()
		if len(logmap) == 0 {
			logmaplock.Unlock()
			break
		}

		var lines []string
		if len(logmap) < 100 {
			lines = logmap
			logmap = nil
		} else {
			lines = logmap[0:100]
			logmap = logmap[100:]
		}
		logmaplock.Unlock()
		if len(lines) > 0 {
			sendLog(lines)
		}
	}
}

func sendLog(lines []string) {
	body := ""
	for _, line := range lines {
		body += base64.StdEncoding.EncodeToString([]byte(line)) + "\n"
	}

	buff := bytes.NewBuffer([]byte(body))
	// retry max 10
	for i := 0; i < 10; i++ {
		resp, err := http.Post(logServerHost+"/collect/?format=base64&secret="+logServerSecret, "text/plain", buff)
		if err != nil {
			fmt.Println("LOG ERR", err.Error(), "RETRY")
			fmt.Println("STREAM LOG ERR 28340539", err.Error(), ", retrying in 10 seconds...")
			time.Sleep(10 * time.Second)
			continue
		}
		if resp.Body != nil {
			resp.Body.Close()
		}
		if resp.StatusCode == 200 {
			break
		}
		fmt.Println("STREAM LOG ERR 19234854", resp.StatusCode, ", retrying in 5 seconds...")
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}
