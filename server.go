package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var errServerHost string
var errServerDomain string
var errServerSecret string
var hostname string

func init() {
	errServerHost = os.Getenv("ERROR_SERVER_HOST")
	if errServerHost == "" {
		errServerHost = "https://track.sbz.vn"
	}
	errServerDomain = os.Getenv("ERROR_SERVER_DOMAIN")
	errServerSecret = os.Getenv("ERROR_SERVER_SECRET")
	hostname, _ = os.Hostname()
	if errServerSecret != "" {
		go flush()
	}
}

var metricmaplock = &sync.Mutex{}
var metricmap = make(map[string]interface{})
var metricmapcount = make(map[string]int)

func flush() {
	// flush periodically in 5s
	for {
		time.Sleep(5 * time.Second)

		metricmaplock.Lock()
		metricmapcopy := make(map[string]any)
		for k, v := range metricmap {
			metricmapcopy[k] = v
		}
		metricmap = make(map[string]any)

		metricmapcountcopy := make(map[string]int)
		for k, v := range metricmapcount {
			metricmapcountcopy[k] = v
		}
		metricmapcount = make(map[string]int)
		metricmaplock.Unlock()

		for metric, theerr := range metricmapcopy {
			count := metricmapcountcopy[metric]
			b, _ := json.Marshal(theerr)

			// retry
			for {
				resp, err := http.Post(errServerHost+"/collects/?type=counter&secret="+errServerSecret+
					"&domain="+errServerDomain+
					"&metric="+metric+
					"&count="+strconv.Itoa(count), "application/json", bytes.NewBuffer(b))
				if err != nil {
					fmt.Println("METRIC ERR", err.Error(), "RETRY")
					time.Sleep(10 * time.Second)
					continue
				}

				if resp.Body != nil {
					resp.Body.Close()
				}
				if resp.StatusCode == 200 || resp.StatusCode >= 400 && resp.StatusCode < 500 {
					break
				}
				fmt.Println("METRIC ERR", resp.StatusCode, "RETRY IN 10sec")
				time.Sleep(10 * time.Second)
			}
		}
	}
}
