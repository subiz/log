package log

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
)

type M map[string]interface{}

type E string

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_missing_resource E = "missing_resource"
const E_access_deny E = "access_deny"
const E_internal E = "internal"
const E_database_error E = "database_error"
const E_file_system_error E = "file_system_error"
const E_transform_data E = "transform_data" // json payload is broken
const E_locked_user E = "locked_user"
const E_unauthorized E = "unauthorized"
const E_wrong_password E = "wrong_password"
const E_user_is_banned E = "user_is_banned"
const E_wrong_signature E = "wrong_signature"

func ErrInvalidInput(base error, required_fields []string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["required_fields"] = required_fields
	return Error(base, field, E_invalid_input)
}

func ErrServer(base error, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(base, field, E_internal)
}

func ErrDB(base error, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(base, field, E_internal, E_database_error)
}

func ErrData(base error, payload []byte, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["size"] = len(payload)
	field["payload"] = string(payload[:200])
	return Error(base, field, E_internal, E_transform_data)
}

func ErrFS(base error, path string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["path"] = path
	return Error(base, field, E_internal, E_file_system_error)
}

func ErrLockedUser(userid string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return Error(nil, field, E_locked_user, E_access_deny)
}

func ErrAccessDeny(userid string, requiredPerm string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return Error(nil, field, E_access_deny)
}

func ErrWrongPassword(fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_wrong_password)
}

func ErrUserIsBanned(accid, userid string, internal_message string, fields ...M) error {
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

func ErrMissing(id, typ string, fields ...M) error {
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

func Error(err error, field M, codes ...E) error {
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

		outerr.Description += string(b) + " " // backward compatitle, remove in future
	}

	if err != nil {
		outerr.XHidden["root"] = err.Error()
	}

	stack, funcname := getStack(0)
	if len(codes) > 0 {
		msg, has := ErrorTable[codes[0]]
		if has {
			outerr.Message = map[string]string{
				"En_US": formatString(msg["en_US"], field),
				"Vi_VN": formatString(msg["vi_VN"], field),
			}
		}
	}

	outerr.Stack = stack // backward compatible, remove in future

	errid := strconv.Itoa(int(crc32.ChecksumIEEE([]byte(funcname + "/" + outerr.Code))))
	outerr.Number = errid
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
	return outerr
}

func OverrideErrorTable(errtable map[E]H) {
	ErrorTable = errtable
}

type AError struct {
	Description string            `json:"description,omitempty"` // remove, prefer i18n message
	Class       int32             `json:"class,omitempty"`       // remove http-code, should be derived from code
	Stack       string            `json:"stack,omitempty"`       // remove
	Code        string            `json:"code,omitempty"`        // should be general database_error, access_deny
	Number      string            `json:"number,omitempty"`      // unique, or hash of stack 4930543478 for grouping error
	Fields      map[string]string `json:"fields,omitempty"`
	XHidden     map[string]string `json:"_hidden,omitempty" `
	Message     map[string]string `json:"message,omitempty"`
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
