package log

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
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
		flushLog()
	}
}

var logmaplock = &sync.Mutex{}
var logmap = []string{}

func flushLog() {
	// flush periodically in 5s
	for {
		time.Sleep(5 * time.Second)
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
			time.Sleep(10 * time.Second)
			continue
		}
		if resp.Body != nil {
			resp.Body.Close()
		}
		if resp.StatusCode == 200 {
			break
		}
		fmt.Println("METRIC ERR", resp.StatusCode, "RETRY IN 10sec")
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}
