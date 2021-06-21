## log
Dead simple logging package for streamming log to syslog

## Usage
Just import and use it
No configuration needed
No new terms to learn.

```go
package main

import "github.com/subiz/log"

func main() {
	log.Info("ac123", "this is a log message", "cid", "cs123")

	var err error
	log.Err("ac123", err, "this is an err", "cid", "cs123")
}

```

Use Info to log info, use Err to log error. Obviously!.
Pass account_id to the first parameter to partition the log stream. It helps you grep log more effeciently.
Thats it.
