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

	"github.com/subiz/header"
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
	go flush()
}

var metricmaplock = &sync.Mutex{}
var metricmap = make(map[string]*header.Event)
var metricmapcount = make(map[string]int)

func flush() {
	// flush periodically in 10s
	for {
		metricmaplock.Lock()
		metricmapcopy := make(map[string]*header.Event)
		for k, v := range metricmap {
			metricmapcopy[k] = v
		}
		metricmap = make(map[string]*header.Event)

		metricmapcountcopy := make(map[string]int)
		for k, v := range metricmapcount {
			metricmapcountcopy[k] = v
		}
		metricmapcount = make(map[string]int)
		metricmaplock.Unlock()

		for metric, theerr := range metricmapcopy {
			count := metricmapcountcopy[metric]
			b, _ := json.Marshal(theerr)
			resp, err := http.Post(errServerHost+"/collects/?type=counter&secret="+errServerSecret+
				"&domain="+errServerDomain+
				"&metric="+metric+
				"&count="+strconv.Itoa(count), "application/json", bytes.NewBuffer(b))
			if err != nil {
				fmt.Println("METRIC ERR", err.Error())
				continue
			}
			if resp != nil {
				if resp.StatusCode != 200 {
					fmt.Println("METRIC ERR", resp.StatusCode)
				}
				if resp.Body != nil {
					resp.Body.Close()
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
