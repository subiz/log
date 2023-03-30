package log

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"strconv"
	"strings"

	"github.com/subiz/header"
)

type M map[string]interface{}

type E string

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_not_found E = "not_found"
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

func ErrNotFound(id, typ string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["type"] = typ
	field["id"] = id
	return Error(nil, field, E_not_found)
}

func IsErr(err error, code E) bool {
	myerr, ok := err.(*header.Error)
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
		mye, ok := err.(*header.Error)
		if !ok {
			// casting to err failed
			// dont give up yet, fallback to json
			errstr := err.Error()
			if strings.HasPrefix(errstr, "#ERR ") {
				roote := &header.Error{}
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

	codestr := ""
	for i, code := range codes {
		if i == 0 {
			codestr = string(code)
			continue
		}
		codestr += "," + string(code)
	}
	// access_deny,locked_user
	outerr := &header.Error{Code: codestr}
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
	}

	if err != nil {
		outerr.XHidden["root"] = err.Error()
	}

	stack, funcname := getStack(0)
	if len(codes) > 0 {
		msg, has := ErrorTable[codes[0]]
		if has {
			outerr.Message = &header.I18NString{
				En_US: formatString(msg["en_US"], field),
				Vi_VN: formatString(msg["vi_VN"], field),
			}
		}
	}

	errid := strconv.Itoa(int(crc32.ChecksumIEEE([]byte(funcname))))
	outerr.Number = errid
	outerr.XHidden["stack"] = stack
	outerr.XHidden["server_name"] = hostname

	if errServerDomain != "" {
		metricmaplock.Lock()
		metricmap[errid] = &header.Event{AccountId: outerr.XHidden["account_id"], UserId: outerr.XHidden["user_id"], Data: &header.Data{Error: outerr}}
		metricmapcount[errid]++
		metricmaplock.Unlock()
	}
	return outerr
}
