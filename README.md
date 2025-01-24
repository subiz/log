## log
Open Telemetry logging

## Usage
```go
package main

import "github.com/subiz/log"

func main() {
	log.Init(APIKEY, "", nil)
	defer log.Shutdown() // prevent missing log when crash

	log.Info("ac123", "this is a log message", "cid", "cs123")

	var err error
	log.Err("ac123", err, "this is an err", "cid", "cs123")

	err := log.Error(ctx, baseerr, "key1", "value1", "key2", 123)

	log.InfoContext(ctx, "message", "key1", "value1", "key2", 123)
}

```

Use Info to log info, use Err to log error.
Pass account_id to the first parameter to partition the log stream. It helps you grep log more effeciently.

```go
package main

import (
    "context"
    "github.com/subiz/log"
)

func parent(ctx context.Context) {
    ctx, span := log.Start(ctx, "the-parent")
    defer span.End()

    log.SetAttributes(span, "isTrue", true, "stringAttr", "hi!")
	// same as
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "hi!"))

    child(ctx)
}

func child(ctx context.Context) {
    ctx, span := log.Start(ctx, "the-child")
    defer span.End()

	log.InfoContext(ctx, "hello", "11111111", "22222222", "333", 4)
}

func main() {
    defer log.Shutdown()

    ctx, span := log.Start(context.Background(), "the-main")
	defer span.End()

	parent(ctx)
}

```
Output:
```bash
2025/01/24 17:18:42 INFO hello world user-id=string user-count=1995 _stacktrace="/home/thanh/src/log/trace_test.go:25 | /home/thanh/src/log/trace_test.go:18 | /home/thanh/src/log/trace_test.go:34"
```

Traces:
```json
{
  "Name": "the-child",
  "SpanContext": {
    "TraceID": "187505507470af3a307e565e0a0da313",
    "SpanID": "5593d09ec38351f0",
    "TraceFlags": "01",
    "TraceState": "",
    "Remote": false
  },
  "Parent": {
    "TraceID": "187505507470af3a307e565e0a0da313",
    "SpanID": "dc01533769632340",
    "TraceFlags": "01",
    "TraceState": "",
    "Remote": false
  },
  "SpanKind": 1,
  "StartTime": "2025-01-24T17:18:42.404704423+07:00",
  "EndTime": "2025-01-24T17:18:42.404801689+07:00",
  "Status": {
    "Code": "Unset",
    "Description": ""
  },
  "Resource": [
    {
      "Key": "host.name",
      "Value": {
        "Type": "STRING",
        "Value": "beast"
      }
    },
    {
      "Key": "service.name",
      "Value": {
        "Type": "STRING",
        "Value": "unknown_service:log.test"
      }
    }
  ]
}

{
  "Name": "the-parent",
  "SpanContext": {
    "TraceID": "187505507470af3a307e565e0a0da313",
    "SpanID": "dc01533769632340",
    "TraceFlags": "01",
    "TraceState": "",
    "Remote": false
  },
  "Parent": {
    "TraceID": "187505507470af3a307e565e0a0da313",
    "SpanID": "71295daa66108522",
    "TraceFlags": "01",
    "TraceState": "",
    "Remote": false
  },
  "SpanKind": 1,
  "StartTime": "2025-01-24T17:18:42.404674242+07:00",
  "EndTime": "2025-01-24T17:18:42.404805926+07:00",
  "Attributes": [
    {
      "Key": "isTrue",
      "Value": {
        "Type": "BOOL",
        "Value": true
      }
    },
    {
      "Key": "stringAttr",
      "Value": {
        "Type": "STRING",
        "Value": "hi!"
      }
    }
  ],
  "Status": {
    "Code": "Unset",
    "Description": ""
  },
  "ChildSpanCount": 1,
  "Resource": [
    {
      "Key": "host.name",
      "Value": {
        "Type": "STRING",
        "Value": "beast"
      }
    }
  ]
}

{
  "Name": "the-main",
  "SpanContext": {
    "TraceID": "187505507470af3a307e565e0a0da313",
    "SpanID": "71295daa66108522",
    "TraceFlags": "01",
    "TraceState": "",
    "Remote": false
  },
  "Parent": {
    "TraceID": "00000000000000000000000000000000",
    "SpanID": "0000000000000000",
    "TraceFlags": "00",
    "TraceState": "",
    "Remote": false
  },
  "SpanKind": 1,
  "StartTime": "2025-01-24T17:18:42.404641201+07:00",
  "EndTime": "2025-01-24T17:18:42.404808277+07:00",
  "Status": {
    "Code": "Unset",
    "Description": ""
  },
  "ChildSpanCount": 1,
  "Resource": [
    {
      "Key": "host.name",
      "Value": {
        "Type": "STRING",
        "Value": "beast"
      }
    }
  ]
}
```
