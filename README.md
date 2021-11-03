## log
Dead simple logging package for streaming log to syslog

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

## Config Dockerfile
In order to stream log from all containers to one centralized server (log-1), you must install `rsyslog` in every container and config it  like bellow.
```
RUN apk update && apk add rsyslog

RUN echo $'\n$FileOwner root\n$FileGroup adm\n$FileCreateMode 0640\n$DirCreateMode 0755\n$Umask 0022\n$WorkDirectory /tmp\n$ActionQueueFileName justlog\n$ActionQueueMaxDiskSpace 1g\n$ActionQueueSaveOnShutdown on\n$ActionQueueType LinkedList\n$ActionResumeRetryCount -1\n$ActionResumeInterval 30\n*.* @log-1:514' >> /etc/rsyslog.conf

```
