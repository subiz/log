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

### Development

#### Thêm một error mới

**Ưu tiên dùng lại hàm lỗi có sẵn. Chỉ tạo hàm mới khi lỗi có ý nghĩa riêng mà các hàm hiện tại không diễn đạt được.** Trước khi thêm, tìm trong `error.go` và `errors.json`, đọc cả tham số, mã lỗi và thông báo của các lỗi tương tự. Khác endpoint, service, tên trường hoặc ID chưa phải lý do để tạo thêm hàm; hãy truyền thông tin đó qua tham số hoặc `fields ...M`.

Ví dụ, nếu chỉ cần báo tài nguyên không tồn tại hoặc giá trị trường không hợp lệ, cân nhắc dùng `ENotFound` / `EMissing` hoặc `EInvalidField`:

```go
return log.ENotFound(recordID, "record", log.M{"account_id": accid})
// Hoặc, tại nơi kiểm tra giá trị trường:
return log.EInvalidField(accid, "status", status)
```

Nếu thực sự cần lỗi mới, thực hiện đủ các bước sau:

1. **Chọn tên theo bản chất lỗi.** Dùng tiền tố `E` và PascalCase, mô tả đối tượng và trạng thái/nguyên nhân, chẳng hạn `ERecordTypeNotEmpty`, hoặc `EInvalid<Field>` cho giá trị không hợp lệ. Tránh tên chung chung như `ECreateFailed`, tên gắn với endpoint hay chi tiết triển khai. Tên cần đủ rõ để nơi khác có thể tìm và dùng lại. Mã lỗi dùng `snake_case`, hằng số tương ứng là `E_<mã_lỗi>`; ví dụ `ERecordTypeNotEmpty` → `record_type_not_empty` → `E_record_type_not_empty`.

2. **Thêm hằng số và hàm vào `error.go`.** Hàm trả về `*AError`, nhận các tham số cần thiết và `fields ...M` để bổ sung ngữ cảnh. Làm theo mẫu của hàm tương tự; khi gọi `Error3`, đặt mã lỗi cụ thể trước, rồi đến các mã phân loại có sẵn phù hợp. Mã đầu tiên được dùng để lấy thông báo trong `ErrorTable`. Ví dụ dưới đây là mẫu đã có trong dự án, không cần thêm lại:

   ```go
   const E_record_type_not_empty E = "record_type_not_empty"

   func ERecordTypeNotEmpty(accid, recordTypeID string, fields ...M) *AError {
       field := M{}
       if len(fields) > 0 && fields[0] != nil {
           field = fields[0]
       }
       field["record_type_id"] = recordTypeID
       return Error3(accid, nil, field, E_record_type_not_empty, E_invalid_input)
   }
   ```

3. **Thêm key vào `errors.json`.** Key phải trùng chính xác với giá trị mã lỗi trong `error.go`, có đủ `vi_VN` và `en_US`. Thông báo cần dễ hiểu với người dùng và hướng dẫn cách khắc phục nếu có. Ví dụ key tương ứng với hàm trên:

   ```json
   "record_type_not_empty": {
     "vi_VN": "Không thể lưu trữ loại bản ghi vì vẫn còn bản ghi chưa xoá. Hãy xoá các bản ghi trước rồi thử lại.",
     "en_US": "This record type cannot be archived while it still contains non-deleted records. Delete those records and try again."
   }
   ```

4. **Chạy build từ thư mục gốc của repository** sau mỗi lần sửa `errors.json`:

   ```sh
   node build.js
   ```

   Có thể dùng `npm run build` tương đương. Lệnh này sinh lại `error_table.go`; không tự sinh hằng số hoặc hàm trong `error.go`. Không sửa trực tiếp `error_table.go`. Kiểm tra diff và commit cùng các thay đổi trong `error.go`, `errors.json` và `error_table.go`.
