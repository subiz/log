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

#### Tra cứu hàm lỗi có sẵn

Ưu tiên chọn hàm theo ý nghĩa lỗi trong bảng dưới đây trước khi tạo hàm mới. Đây là các hàm phổ biến; xem `error.go` để tìm thêm các lỗi nghiệp vụ cụ thể. Cột mã lỗi liệt kê mã chính trước, rồi đến các mã phân loại.

Các hàm trong bảng đều nhận thêm `fields ...M` ở cuối để bổ sung ngữ cảnh, ví dụ `log.M{"account_id": accid, "record_id": recordID}`. Các hàm chỉ đọc map đầu tiên và sao chép map trước khi thêm thuộc tính; chúng không sửa map của bên gọi. Các giá trị bên trong map không được sao chép sâu. Thuộc tính có tiền tố `_` dành cho dữ liệu nội bộ và không được đưa vào `ToJSON()` trả cho người dùng.

| Nhóm | Hàm và tham số chính | Khi nào dùng | Mã lỗi |
| --- | --- | --- | --- |
| Validation | `EInvalidField(accid, name, value string)` | Giá trị của một trường không hợp lệ | `invalid_field`, `invalid_input` |
| Validation | `EInvalidInput(base error, required_fields []string, internal_message string)` | Đầu vào không đáp ứng yêu cầu, cần nêu các trường bắt buộc | `invalid_input` |
| Validation | `EInvalidInputFormat(base error, fieldname, currentvalue, msg string)` | Giá trị sai định dạng | `invalid_input_format`, `invalid_input` |
| Validation | `EInvalidUrl(accid, url string)` | URL không hợp lệ | `invalid_url`, `invalid_input` |
| Validation | `EPayloadTooLarge(curSize, maxSize int64)` | Dữ liệu vượt kích thước cho phép | `payload_too_large`, `invalid_input` |
| Giới hạn | `ELimitExceeded(capacity int64)` | Vượt hạn mức | `limit_exceeded`, `invalid_input` |
| Giới hạn | `ERateLimit()` | Gửi yêu cầu quá nhanh | `rate_limit`, `limit_exceeded`, `invalid_input` |
| Tài nguyên | `ENotFound(id, typ string)` / `EMissing(id, typ string)` | Không tìm thấy tài nguyên; hai hàm dùng cùng mã lỗi | `missing_resource` |
| Tài nguyên | `EInactive(id, typ string)` | Tài nguyên không hoạt động | `inactive_resource` |
| Quyền truy cập | `EDeny(userid, requiredPerm string)` | Không đủ quyền truy cập | `access_deny` |
| Quyền truy cập | `EExpiredAccessToken()` | Access token đã hết hạn | `access_deny`, `access_token_expired` |
| Quyền truy cập | `EAccountLocked(accid string)` | Tài khoản bị khoá | `locked_account` |
| Hệ thống | `EServer(base error)` | Lỗi nội bộ | `internal` |
| Hệ thống | `ERetry(base error)` | Lỗi nội bộ có thể thử lại | `internal`, `retryable` |
| Hệ thống | `ETimeout()` | Xử lý quá thời gian, có thể thử lại | `timeout`, `internal`, `retryable` |
| Hệ thống | `EServiceUnavailable(err error)` | Dịch vụ tạm thời không sẵn sàng | `service_unavailable`, `internal`, `retryable` |
| Hệ thống | `EInternalConnect(base error)` | Lỗi kết nối nội bộ | `internal_connection`, `internal` |
| Hệ thống | `EFS(base error, path string)` | Lỗi hệ thống file | `file_system_error`, `internal` |
| Provider | `EProvider(base error, external_service, action string)` | Lỗi từ dịch vụ bên ngoài | `provider_failed`, `internal` |

Các hàm `E...` trên tạo lỗi **và ghi log** qua `Error3`. Nếu chỉ muốn tạo hoặc bọc lỗi, dùng `NewError(base, fields, codes...)` hoặc `NewError2(ctx, base, codes, key, value, ...)`. `NewError2` nhận các cặp key/value và gắn thông tin trace từ context khi tạo lỗi mới; khi bọc `AError` có sẵn, giữ thông tin trace của lỗi gốc.

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
       field := copyErrorFields(fields...)
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

   Có thể dùng `npm run build` tương đương. Build kiểm tra mã lỗi theo `snake_case`, đủ hai bản dịch không rỗng và bộ placeholder giống nhau giữa hai ngôn ngữ (ví dụ cùng dùng `{name}`). Build tự escape ký tự đặc biệt khi sinh chuỗi Go; nếu dữ liệu không hợp lệ, lệnh báo key cần sửa và không ghi đè bảng hiện tại.

   Lệnh này sinh lại `error_table.go`; không tự sinh hằng số hoặc hàm trong `error.go`. Không sửa trực tiếp `error_table.go`. Kiểm tra diff và commit cùng các thay đổi trong `error.go`, `errors.json` và `error_table.go`.
