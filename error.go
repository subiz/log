// TODO: join stacktrace through grpc

package log

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
)

// special field
// + code
// + payload_string
type M map[string]interface{}

type E string

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_invalid_domain E = "invalid_domain"
const E_missing_resource E = "missing_resource"
const E_access_deny E = "access_deny"
const E_missing_id E = "missing_id"
const E_internal E = "internal"
const E_not_a_conversation_member E = "not_a_conversation_member"
const E_file_system_error E = "file_system_error"
const E_transform_data E = "transform_data" // json payload is broken
const E_data_corrupted E = "data_corrupted" // json payload is broken
const E_locked_user E = "locked_user"
const E_unauthorized E = "unauthorized"
const E_wrong_password E = "wrong_password"
const E_user_is_banned E = "user_is_banned"
const E_user_is_unsubscribed E = "user_is_unsubscribed"
const E_wrong_signature E = "wrong_signature"
const E_access_token_expired E = "access_token_expired"
const E_locked_account E = "locked_account"
const E_locked_agent E = "locked_agent"
const E_internal_connection E = "internal_connection"
const E_provider_failed E = "provider_failed"
const E_provider_data_mismatched E = "provider_data_mismatched"
const E_payload_too_large E = "payload_too_large"
const E_limit_exceeded E = "limit_exceeded"
const E_service_unavailable E = "service_unavailable"
func EServiceUnavailable(err error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(err, field, E_service_unavailable, E_internal)
}

func EPayloadTooLarge(curSize int64, maxSize int64, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["current_size"] = curSize
	field["maximum_size"] = maxSize
	return Error(nil, field, E_payload_too_large, E_invalid_input)
}

func ELimitExceeded(cur int64, capacity int64, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["current"] = cur
	field["capacity"] = capacity
	return Error(nil, field, E_limit_exceeded, E_invalid_input)
}

func EMissingId(typ string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_missing_id, E_invalid_input)
}

func EInvalidInput(base error, required_fields []string, internal_message string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["required_fields"] = required_fields
	return Error(base, field, E_invalid_input)
}

func EAgentLocked(accid, agentid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["agent_id"] = agentid
	return Error(nil, field, E_locked_agent, E_internal)
}

func EAccountLocked(accid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	return Error(nil, field, E_locked_account, E_internal)
}

func EServer(base error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(base, field, E_internal)
}

func EProvider(base error, external_service, action string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["external_service"] = external_service
	field["action"] = action
	return Error(base, field, E_provider_failed, E_internal)
}

func EData(base error, payload []byte, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["size"] = len(payload)
	field["payload"] = Substring(string(payload), 0, 200)
	return Error(base, field, E_internal, E_transform_data, E_data_corrupted)
}

func EInternalConnect(base error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(base, field, E_internal_connection, E_internal)
}

func EFS(base error, path string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["path"] = path
	return Error(base, field, E_file_system_error, E_internal)
}

func ELockedUser(userid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return Error(nil, field, E_locked_user, E_access_deny)
}

func EDeny(userid string, requiredPerm string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return Error(nil, field, E_access_deny)
}

func EWrongPassword(fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_wrong_password)
}

func EUnsub(accid, userid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["user_id"] = userid
	return Error(nil, field, E_user_is_unsubscribed, E_invalid_input)
}

func EBanned(accid, userid string, internal_message string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["user_id"] = userid
	return Error(nil, field, E_user_is_banned)
}

func ErrContext(ctx context.Context, err error) error {
	return err
}

func EMissing(id, typ string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["type"] = typ
	field["id"] = id
	return Error(nil, field, E_missing_resource)
}

func IsErr(err error, code string) bool {
	myerr, ok := err.(*AError)
	if !ok {
		return false
	}
	codes := strings.Split(myerr.Code, ",")
	codestring := string(code)
	for _, c := range codes {
		if codestring == c {
			return true
		}
	}
	return false
}

func Error(err error, field M, codes ...E) *AError {
	if err != nil {
		mye, ok := err.(*AError)
		if !ok {
			// casting to err failed
			// dont give up yet, fallback to json
			errstr := err.Error()
			if strings.HasPrefix(errstr, "#ERR ") {
				roote := &AError{}
				if er := json.Unmarshal([]byte(errstr[len("#ERR "):]), roote); er == nil {
					if roote.Code != "" && roote.Class != 0 { // valid err
						mye = roote
					}
				}
			}
		}

		// our error
		if mye != nil {
			for key, value := range field {
				if key == "" || key == "__skip_stack" {
					continue
				}
				b, _ := json.Marshal(value)
				if key[0] == '_' {
					if mye.XHidden == nil {
						mye.XHidden = map[string]string{}
					}
					mye.XHidden[key[1:]] = string(b)
				} else {
					if mye.Fields == nil {
						mye.Fields = map[string]string{}
					}
					mye.Fields[key] = string(b)
				}
			}
			return mye
		}
	}

	// access_deny,locked_user
	outerr := &AError{}
	// backward compatible, remove in future
	outerr.Class = 400
	codestr := ""
	for i, code := range codes {
		if code == E_internal {
			outerr.Class = 500
		}

		if i == 0 {
			codestr = string(code)
			continue
		}
		codestr += "," + string(code)
	}
	outerr.Code = codestr
	outerr.Fields = map[string]string{}
	outerr.XHidden = map[string]string{}

	skipstack := 0
	if field["__skip_stack"] != nil {
		skipstack = interfaceToInt(field["__skip_stack"])
		delete(field, "__skip_stack")
	}

	for key, value := range field {
		if key == "" {
			continue
		}
		b, _ := json.Marshal(value)
		if key[0] == '_' {
			outerr.XHidden[key[1:]] = string(b)
		} else {
			outerr.Fields[key] = string(b)
		}
	}

	if err != nil {
		outerr.XHidden["root"] = err.Error()
	}

	stack, funcname := getStack(skipstack)
	if len(codes) > 0 {
		msg, has := ErrorTable[codes[0]]
		if has {
			outerr.Message = map[string]string{
				"En_US": formatString(msg["en_US"], field),
				"Vi_VN": formatString(msg["vi_VN"], field),
			}
		}
	}

	// override message
	if field["_message"] != nil {
		imessage := field["_message"]
		message, ok := imessage.(map[string]string)
		if ok {
			for code, msg := range message {
				outerr.Message[code] = msg
			}
		}
	}

	errid := strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(funcname+"/"+outerr.Code))), 16)
	outerr.Number = errid
	if funcname != "" {
		outerr.XHidden["function_name"] = funcname
	}

	if field["_function_name"] != nil {
		if funcname, _ := field["_function_name"].(string); funcname != "" {
			outerr.XHidden["function_name"] = funcname
		}
	}
	outerr.XHidden["stack"] = stack
	outerr.XHidden["server_name"] = hostname

	if errServerDomain != "" {
		metricmaplock.Lock()
		metricmap[errid] = map[string]any{
			"account_id": outerr.XHidden["account_id"],
			"created":    time.Now().UnixMilli(),
			"type":       "error_" + errid,
			"user_id":    outerr.XHidden["user_id"],
			"data":       map[string]any{"error": outerr},
		}
		metricmapcount[errid]++
		metricmaplock.Unlock()
	}

	if errVerbose != "" {
		// try to get accid
		var accid string
		message := outerr.Message["Vi_VN"]
		if message == "" {
			message = outerr.Message["En_US"]
		}

		m := fmt.Sprintf("ERR %s [%s]. %v %v", message, outerr.Code, outerr.Fields, outerr.XHidden)
		w.writeAndRetry(accid, LOG_ERR, m)
	}

	return outerr
}

func WrapStack(err error, skip int) error {
	if err == nil {
		return nil
	}

	mye, ok := err.(*AError)
	if !ok {
		// casting to err failed
		// dont give up yet, fallback to json
		errstr := err.Error()
		if strings.HasPrefix(errstr, "#ERR ") {
			roote := &AError{}
			if er := json.Unmarshal([]byte(errstr[len("#ERR "):]), roote); er == nil {
				if roote.Code != "" && roote.Class != 0 { // valid err
					mye = roote
				}
			}
		}
	}

	// not our error
	if mye == nil {
		return err
	}

	stack, _ := getStack(skip)
	if mye.XHidden == nil {
		mye.XHidden = map[string]string{}
	}
	mye.XHidden["stack"] = mye.XHidden["stack"] + "\n--\n" + stack
	return mye
}

func OverrideErrorTable(errtable map[E]H) {
	ErrorTable = errtable
}

type AError struct {
	Class   int32             `json:"class,omitempty"`  // remove http-code, should be derived from code
	Code    string            `json:"code,omitempty"`   // should be general database_error, access_deny
	Number  string            `json:"number,omitempty"` // unique, or hash of stack 4930543478 for grouping error
	Fields  map[string]string `json:"fields,omitempty"`
	XHidden map[string]string `json:"_hidden,omitempty" `
	Message map[string]string `json:"message,omitempty"`
}

// Error returns string representation of an Error
func (e *AError) Error() string {
	if e == nil {
		return ""
	}

	b, _ := json.Marshal(e)
	return "#ERR " + string(b)
}

// FromString unmarshal an error string to *Error
func FromString(err string) *AError {
	if !strings.HasPrefix(err, "#ERR ") {
		return nil
	}
	e := &AError{}
	if er := json.Unmarshal([]byte(err[len("#ERR "):]), e); er != nil {
		return nil
	}
	return e
}

func Substring(s string, start int, end int) string {
	if start == 0 && end >= len(s) {
		return s
	}

	start_str_idx := 0
	i := 0
	for j := range s {
		if i == start {
			start_str_idx = j
		}
		if i == end {
			return s[start_str_idx:j]
		}
		i++
	}
	return s[start_str_idx:]
}

// var ISOTABLE = crc64.MakeTable(crc64.ISO)

func interfaceToInt(in interface{}) int {
	var number int
	switch v := in.(type) {
	case string:
		number, _ = strconv.Atoi(v)
	case int:
		number = v
	case int64:
		number = int(v)
	case int32:
		number = int(v)
	case uint64:
		number = int(v)
	case uint32:
		number = int(v)
	case float32:
		number = int(v)
	case float64:
		number = int(v)
	}
	return number
}
