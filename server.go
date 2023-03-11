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

var serverEnv string
var hostname string

func init() {
	serverEnv = os.Getenv("LOG_SERVER_ENV")
	hostname, _ = os.Hostname()
	go flush()
}

var metricmaplock = &sync.Mutex{}
var metricmap = make(map[int64]*header.Event)
var metricmapcount = make(map[int64]int)

func flush() {
	// flush periodically in 10s
	for {
		start := time.Now()
		metricmaplock.Lock()
		metricmapcopy := make(map[int64]*header.Event)
		for k, v := range metricmap {
			metricmapcopy[k] = v
		}
		metricmap = make(map[int64]*header.Event)

		metricmapcountcopy := make(map[int64]int)
		for k, v := range metricmapcount {
			metricmapcountcopy[k] = v
		}
		metricmapcount = make(map[int64]int)
		metricmaplock.Unlock()

		for metric, theerr := range metricmapcopy {
			count := metricmapcountcopy[metric]
			b, _ := json.Marshal(theerr)
			resp, err := http.Post("https://"+serverEnv+"/collects/?type=counter&secret=iamnobot&count="+strconv.Itoa(count), "application/json", bytes.NewBuffer(b))
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

		if len(metricmapcopy) > 0 {
			fmt.Println("METRIC FLUSHED:", len(metricmapcopy), "in", time.Since(start))
		}
		time.Sleep(10 * time.Second)
	}
}
