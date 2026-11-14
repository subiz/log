# log

Thư viện ghi log và xử lý lỗi dùng chung cho Subiz. Go ghi log ra console và gửi
qua OTLP HTTP tại `/v1/logs`; package JavaScript cung cấp thông báo lỗi và hàm
thay thế placeholder. Thư viện không tạo hoặc xuất distributed trace.

## Cấu hình

Đặt biến môi trường trước khi khởi động ứng dụng:

| Biến | Ý nghĩa | Mặc định |
| --- | --- | --- |
| `LOG_SERVER_HOST` | URL gốc của máy chủ nhận log | `https://log.subiz.net` |
| `LOG_SERVER_SECRET` | Secret gửi trong HTTP header `secret`; để trống thì chỉ ghi console | Trống |
| `LOG_HOSTNAME` | Ghi đè tên máy chủ trong log | `os.Hostname()` |

`service.name` lấy phần hostname trước dấu `-` đầu tiên. Ví dụ,
`zaloperson-75fb677958-65dz6` có tên dịch vụ là `zaloperson`.
`OTEL_SDK_DISABLED` không tắt luồng gửi log của package này.

Các API dùng chung hàng đợi OTLP tối đa 100.000 bản ghi, gửi tối đa 100 bản ghi
mỗi batch, theo chu kỳ 2 giây hoặc sớm hơn khi đủ batch.

## Ghi log

```go
package main

import (
    "context"
    "errors"

    "github.com/subiz/log"
)

func main() {
    defer log.Shutdown()
    ctx := context.Background()

    log.InfoContext(ctx, "request received", "account_id", "ac123", "cid", "cs123")
    log.WarnContext(ctx, "request is slow", "account_id", "ac123")
    log.Error(ctx, errors.New("database connection refused"), "internal,retryable",
        "account_id", "ac123")
}
```

Truyền `account_id` trực tiếp trong thuộc tính để phân loại log theo tài khoản.
Các API nhận context dùng console handler hiện tại của `slog.Default()` và tuân
theo mức log của handler đó (mặc định là Info). `Error` ghi cả lỗi có cấu trúc,
nguyên nhân và stack ra console và OTLP, đồng thời trả về `*AError`.

API cũ `Info(accountID, ...)` và `Err(accountID, err, ...)` vẫn dùng được, giữ
định dạng console riêng và gửi qua cùng hàng đợi OTLP. `Err` bỏ qua lỗi nil.
`Track(ctx, code, ...)` ghi một sự kiện log với tiền tố `TRK-`.

`Start` trả về span không ghi dữ liệu để tương thích với caller cũ. Các hàm
`SetAttributes` và `SetSpanAttributes` không thay thế thuộc tính log; hãy truyền
thuộc tính vào từng lời gọi ghi log.

## Dừng ứng dụng

Thư viện không tự bắt SIGINT/SIGTERM. Ứng dụng cần dừng và chờ các worker hoàn
thành trước khi gọi shutdown để gửi hết log đang chờ:

- `Shutdown()` chờ tối đa 10 giây và in lỗi ra stderr.
- `ShutdownContext(ctx)` dùng thời hạn của caller và trả về lỗi.

Nếu hết thời gian, một số log có thể chưa được gửi; gọi shutdown lần nữa không
thử lại. Dùng context mới cho shutdown vì context nhận signal đã bị huỷ:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/subiz/log"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    log.InfoContext(ctx, "service started")
    <-ctx.Done()
    stop() // Khôi phục xử lý signal mặc định trong lúc shutdown.
    // Dừng và chờ các worker của ứng dụng tại đây.

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := log.ShutdownContext(shutdownCtx); err != nil {
        fmt.Fprintln(os.Stderr, "log shutdown:", err)
    }
}
```

## Xử lý lỗi

Các hàm `E...` dưới đây tạo lỗi **và ghi log** qua `Error3`. Nếu chỉ muốn tạo hoặc
bọc lỗi, dùng `NewError(base, fields, codes...)` hoặc
`NewError2(ctx, base, codes, key, value, ...)`.

`IsErr` và các hàm tạo lỗi nhận được `AError` trong chuỗi bọc bằng `%w`. `Unwrap`
giữ nguyên nhân cho `errors.Is`/`errors.As` trong cùng process; nguyên nhân dạng
Go không được serialize. Khi bọc lỗi:

- Mã lỗi được gộp không trùng; thêm `internal` cập nhật `Class` thành 500.
- Thêm `retryable` đổi tiền tố số lỗi từ `SBQ-` sang `SBR-`.
- ID, phần hash, thông báo và trace của lỗi gốc được giữ lại. `NewError2` chỉ lấy
  trace từ context khi tạo lỗi mới.

### Tra cứu hàm có sẵn

Chọn hàm theo ý nghĩa lỗi trước khi tạo hàm mới. Bảng liệt kê các hàm phổ biến;
xem [error.go](error.go) để tìm lỗi nghiệp vụ cụ thể. Mã chính đứng trước các mã
phân loại.

Các hàm đều nhận thêm `fields ...M`, ví dụ
`log.M{"account_id": accid, "record_id": recordID}`. Chỉ map đầu tiên được dùng
và được sao chép trước khi thêm thuộc tính; giá trị bên trong không được sao
chép sâu. Thuộc tính có tiền tố `_` vẫn được ghi log nhưng không xuất hiện trong
`ToJSON()` trả cho người dùng.

| Nhóm | Hàm và tham số chính | Khi nào dùng | Mã lỗi |
| --- | --- | --- | --- |
| Đầu vào | `EInvalidField(accid, name, value string)` | Giá trị của một trường không hợp lệ | `invalid_field`, `invalid_input` |
| Đầu vào | `EInvalidInput(base error, required_fields []string, internal_message string)` | Đầu vào không đáp ứng yêu cầu, cần nêu các trường bắt buộc | `invalid_input` |
| Đầu vào | `EInvalidInputFormat(base error, fieldname, currentvalue, msg string)` | Giá trị sai định dạng | `invalid_input_format`, `invalid_input` |
| Đầu vào | `EInvalidUrl(accid, url string)` | URL không hợp lệ | `invalid_url`, `invalid_input` |
| Đầu vào | `EPayloadTooLarge(curSize, maxSize int64)` | Dữ liệu vượt kích thước cho phép | `payload_too_large`, `invalid_input` |
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
| Dịch vụ ngoài | `EProvider(base error, external_service, action string)` | Lỗi từ dịch vụ bên ngoài | `provider_failed`, `internal` |

## JavaScript

```js
const { messages, format } = require('@subiz/log')

const message = messages.invalid_field.vi_VN
console.log(format(message, { name: 'email', value: 'invalid' }))
```

`messages` chứa bản dịch `vi_VN` và `en_US` từ `errors.json`. `format` thay
placeholder dạng `{name}` bằng giá trị tương ứng. Placeholder thiếu dữ liệu
hoặc chưa đóng dấu `}` được giữ nguyên; `data` có thể bỏ qua hoặc là `null`.

## Phát triển

### Thêm lỗi mới

Ưu tiên dùng lại hàm có sẵn. Khác endpoint, service, tên trường hoặc ID chưa phải
lý do tạo thêm hàm; truyền ngữ cảnh qua tham số hoặc `fields ...M`:

```go
return log.ENotFound(recordID, "record", log.M{"account_id": accid})
```

Nếu cần lỗi có ý nghĩa riêng, kiểm tra [error.go](error.go) và
[errors.json](errors.json), rồi thực hiện các bước sau:

1. **Chọn tên theo bản chất lỗi.** Hàm dùng tiền tố `E` và PascalCase; mã lỗi
   dùng `snake_case`, hằng số là `E_<mã_lỗi>`. Ví dụ: `ERecordTypeNotEmpty`,
   `record_type_not_empty`, `E_record_type_not_empty`. Tránh tên chung chung như
   `ECreateFailed` hoặc tên gắn với endpoint.

2. **Thêm hằng số và hàm vào `error.go`.** Hàm trả về `*AError`, nhận tham số
   cần thiết và `fields ...M`. Sao chép map bằng `copyErrorFields`, gọi `Error3`
   với mã cụ thể trước rồi đến mã phân loại. Mã đầu tiên quyết định thông báo
   trong `ErrorTable`. Mẫu đã có trong repo:

   ```go
   const E_record_type_not_empty E = "record_type_not_empty"

   func ERecordTypeNotEmpty(accid, recordTypeID string, fields ...M) *AError {
       field := copyErrorFields(fields...)
       field["record_type_id"] = recordTypeID
       return Error3(accid, nil, field, E_record_type_not_empty, E_invalid_input)
   }
   ```

3. **Thêm bản dịch vào `errors.json`.** Key trùng với mã lỗi, có đủ `vi_VN` và
   `en_US`, dùng cùng bộ placeholder. Thông báo cần dễ hiểu và hướng dẫn cách
   khắc phục nếu có. Ví dụ entry tương ứng với hàm trên:

   ```json
   {
     "record_type_not_empty": {
       "vi_VN": "Không thể lưu trữ loại bản ghi vì vẫn còn bản ghi chưa xoá. Hãy xoá các bản ghi trước rồi thử lại.",
       "en_US": "This record type cannot be archived while it still contains non-deleted records. Delete those records and try again."
     }
   }
   ```

4. **Sinh lại bảng thông báo** từ thư mục gốc sau mỗi lần sửa `errors.json`:

   ```sh
   npm run build
   ```

   Build kiểm tra mã `snake_case`, bản dịch không rỗng và bộ placeholder giống
   nhau, rồi escape chuỗi và sinh `error_table.go`. Dữ liệu không hợp lệ sẽ báo
   lỗi trước khi ghi đè file. Build không sinh hằng số hoặc hàm trong `error.go`.
   Không sửa trực tiếp `error_table.go`; kiểm tra diff và commit cùng các thay
   đổi trong `error.go`, `errors.json` và `error_table.go`.

### Kiểm tra

```sh
LOG_SERVER_SECRET= go test -race ./...
go vet ./...
npm test
```

Go và JavaScript dùng chung bộ kiểm thử formatter trong
[testdata/format.json](testdata/format.json).
