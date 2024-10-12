package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math/rand"
	"strconv"
	"strings"
)

// special field
// + code
// + payload_string
type M map[string]interface{}

type E string

func (e E) String() string {
	return string(e)
}

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_expired_access_token E = "access_token_expired"
const E_invalid_otp E = "invalid_otp"
const E_invalid_input_format E = "invalid_input_format"
const E_invalid_field E = "invalid_field"
const E_email_taken E = "email_taken"
const E_invalid_domain E = "invalid_domain"
const E_missing_resource E = "missing_resource"
const E_dupplicate_contact_update E = "dupplicate_contact_update"
const E_access_deny E = "access_deny"
const E_missing_id E = "missing_id"
const E_internal E = "internal"
const E_retryable E = "retryable" // -> retryable
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
const E_invite_link_expired E = "invite_link_expired"
const E_locked_account E = "locked_account"
const E_locked_agent E = "locked_agent"
const E_invalid_currency E = "invalid_currency"
const E_internal_connection E = "internal_connection"
const E_provider_failed E = "provider_failed"
const E_provider_data_mismatched E = "provider_data_mismatched"
const E_payload_too_large E = "payload_too_large"
const E_limit_exceeded E = "limit_exceeded"
const E_rate_limit E = "rate_limit"
const E_service_unavailable E = "service_unavailable"
const E_invalid_zalo_token E = "invalid_zalo_token"
const E_invalid_facebook_token E = "invalid_facebook_token"
const E_invalid_google_token E = "invalid_google_token"
const E_insufficient_credit E = "insufficient_credit"
const E_invalid_connection E = "invalid_connection"
const E_invalid_password_length E = "invalid_password_length"
const E_invalid_promotion_code E = "invalid_promotion_code"
const E_conversation_ended E = "conversation_ended"
const E_remote_error E = "remote_error"
const E_invalid_field_size E = "invalid_field_size"
const E_malformed_request E = "malformed_request" // user cannot resolve
const E_invalid_integration E = "invalid_integration"
const E_still_have_open_invoice E = "still_have_open_invoice"
const E_invalid_subscription E = "invalid_subscription"
const E_fb_outside_send_window E = "fb_outside_send_window"
const E_inactive_number E = "inactive_number"
const E_blocked_number E = "blocked_number"
const E_invalid_webhook_url E = "invalid_webhook_url"
const E_password_too_weak E = "password_too_weak"
const E_leaver_is_the_last_one_in_conversation E = "leaver_is_the_last_one_in_conversation"
const E_google_error E = "google_error"
const E_close_public_channel E = "close_public_channel"

func EClosePublicChannel(accid, convoid, channel string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["conversation_id"] = convoid
	field["channel"] = channel
	return Error(nil, field, E_close_public_channel, E_invalid_input)
}

func EPasswordTooWeak(fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_password_too_weak, E_invalid_input)
}

func EGoogle(err error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_google_error, E_internal, E_retryable)
}

func EInvalidPromotionCode(code string, errors []string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["promotion_code"] = code
	field["errors"] = errors
	return Error(nil, field, E_invalid_promotion_code, E_invalid_input)
}

func EInvalidCurrency(accid, cur string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["currency"] = cur
	return Error(nil, field, E_invalid_currency, E_invalid_input)
}

func EStillHaveOpenInvoice(accid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	return Error(nil, field, E_still_have_open_invoice, E_invalid_subscription, E_invalid_input)
}

func ELeaverIsTheLastOneInConvo(accid, convoid, issuerid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["conversation_id"] = convoid
	field["issuer_id"] = issuerid
	return Error(nil, field, E_leaver_is_the_last_one_in_conversation, E_access_deny, E_invalid_input)
}

func ENotAMember(accid, convoid, issuerid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["conversation_id"] = convoid
	field["issuer_id"] = issuerid
	return Error(nil, field, E_not_a_conversation_member, E_access_deny, E_invalid_input)
}

func EInvalidWebhookUrl(accid, webhookurl string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["webhook_url"] = webhookurl
	return Error(nil, field, E_invalid_webhook_url, E_invalid_input)
}

func EInvalidOTP(username, otp string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["username"] = username
	field["otp"] = otp
	return Error(nil, field, E_invalid_otp, E_invalid_input)
}

func EInactiveNumber(accid, number string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["number"] = number
	return Error(nil, field, E_inactive_number, E_invalid_integration, E_invalid_input)
}

func EBlockedNumber(accid, number string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["number"] = number
	return Error(nil, field, E_blocked_number, E_invalid_integration, E_invalid_input)
}

func EInviteLinkExpired(link string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["link"] = link
	return Error(nil, field, E_invite_link_expired, E_invalid_input)
}

func EInvalidIntegration(accid, inteid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["integration_id"] = inteid
	return Error(nil, field, E_invalid_integration, E_invalid_input)
}

func EMalformedRequest(accid, code string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["code"] = code
	return Error(nil, field, E_malformed_request, E_invalid_input)
}

func EConversationEnded(accid, convoid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["conversation_id"] = convoid
	return Error(nil, field, E_conversation_ended, E_invalid_input)
}

func EInvalidPasswordLength(currentLength, requiredLength int, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["current_length"] = currentLength
	field["required_length"] = requiredLength
	return Error(nil, field, E_invalid_password_length, E_invalid_input)
}

func EEmailTaken(email string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["email"] = email
	return Error(nil, field, E_email_taken, E_invalid_input)
}

func EInvalidPollConnection(accid, id string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["connection_id"] = id
	return Error(nil, field, E_invalid_connection, E_invalid_input)
}

func EDeadPollConnection(accid, id string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["connection_id"] = id
	return Error(nil, field, E_invalid_connection, E_invalid_input)
}

func EInvalidField(accid, name, value string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["name"] = name
	field["value"] = value
	return Error(nil, field, E_invalid_field, E_invalid_input)
}

func EDupplicateContactUpdate(accid, emailorphone string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["prop"] = emailorphone
	return Error(nil, field, E_dupplicate_contact_update, E_invalid_input)
}

func ENotEnoughCredit(accid, creditid, creditname, service string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["credit_id"] = creditid
	field["credit_name"] = creditname
	field["service"] = service
	return Error(nil, field, E_insufficient_credit)
}

func EInvalidZaloToken(accid, oaid, oaName string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["oa_id"] = oaid
	field["oa_name"] = oaName
	return Error(nil, field, E_invalid_zalo_token, E_service_unavailable, E_internal)
}

func EInvalidFacebookToken(accid, pageid, pageName string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["page_id"] = pageid
	field["page_name"] = pageName
	return Error(nil, field, E_invalid_facebook_token, E_service_unavailable, E_internal)
}

func EInvalidGoogleToken(accid, locationId, locationName string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["location_id"] = locationId
	field["location_name"] = locationName
	return Error(nil, field, E_invalid_google_token, E_service_unavailable, E_internal)
}

func EServiceUnavailable(err error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(err, field, E_service_unavailable, E_internal, E_retryable)
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

func ELimitExceeded(capacity int64, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["capacity"] = capacity
	return Error(nil, field, E_limit_exceeded, E_invalid_input)
}

func ERateLimit(fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_rate_limit, E_limit_exceeded, E_invalid_input)
}

func EMissingId(typ string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_missing_id, E_invalid_input)
}

func EInvalidInputFormat(base error, fieldname, currentvalue string, msg string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["invalid_field"] = fieldname
	field["invalid_value"] = currentvalue
	field["msg"] = msg
	return Error(base, field, E_invalid_input_format, E_invalid_input)
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
	return Error(nil, field, E_locked_agent)
}

func EAccountLocked(accid string, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	return Error(nil, field, E_locked_account)
}

func ERetry(base error, fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(base, field, E_internal, E_retryable)
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

func EExpiredAccessToken(fields ...M) *AError {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return Error(nil, field, E_access_deny, E_access_token_expired)
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

// NewError creates or wraps a new error
func NewError(err error, field M, codes ...E) *AError {
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
		} else {
			mye = mye.Clone()
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

				codestr := mye.Code
				for _, code := range codes {
					if codestr == "" {
						codestr = string(code)
						continue
					}
					codestr += "," + string(code)
				}
				mye.Code = codestr
			}
			return mye
		}
	}

	outerr := &AError{Id: rand.Int63()}
	// backward compatible, remove in future
	outerr.Class = 400
	codestr := ""
	retryable := false
	for i, code := range codes {
		if code == E_retryable {
			retryable = true
		}
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

	stack, funcname := GetStack(1 + skipstack)
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

	errid := strings.ToUpper(strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(funcname+"/"+outerr.Code))), 16))

	// classified: database error, filesystem error, account access deny
	// SBZ-ER72BFD5F
	// SBZ-EQA2BFD5F
	prefix := "Q"
	if retryable {
		prefix = "R"
	}
	outerr.Number = "SBZ-E" + prefix + errid
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
	return outerr
}

// Error creates an error and report it to server
// If you dont want to report it, use NewError instead
func Error(err error, field M, codes ...E) *AError {
	outerr := NewError(err, field, codes...)
	b, _ := json.Marshal(outerr)
	accid := outerr.XHidden["account_id"]
	w.writeAndRetry(accid, LOG_ERR, string(b))
	return outerr
}

func (ae *AError) ToJSON() string {
	if ae == nil {
		return "null"
	}

	messageb, _ := json.Marshal(ae.Message)
	fieldb, _ := json.Marshal(ae.Fields)
	out := `{"id": ` + strconv.FormatInt(int64(ae.Id), 10) + `,"code":` + fmt.Sprintf("%q", ae.Code) + `,"number":` + fmt.Sprintf("%q", ae.Number) + `,"fields":` + string(fieldb) + `,"message":` + string(messageb) + ``
	return `{"code":` + fmt.Sprintf("%q", ae.Code) + `,"class":` + strconv.FormatInt(int64(ae.Class), 10) + `,"error":` + out + `}}`
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

	stack, _ := GetStack(skip)
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
	Id      int64             `json:"id,omitempty"`
	Class   int32             `json:"class,omitempty"`  // remove http-code, should be derived from code
	Code    string            `json:"code,omitempty"`   // should be general database_error, access_deny
	Number  string            `json:"number,omitempty"` // unique, or hash of stack 4930543478 for grouping error
	Fields  map[string]string `json:"fields,omitempty"`
	XHidden map[string]string `json:"_hidden,omitempty" `
	Message map[string]string `json:"message,omitempty"`
}

func (e *AError) Clone() *AError {
	if e == nil {
		return nil
	}

	clone := &AError{
		Id:      e.Id,
		Class:   e.Class,
		Code:    e.Code,
		Number:  e.Number,
		Message: e.Message,
	}
	if len(e.XHidden) > 0 {
		newXHidden := map[string]string{}
		for k, v := range e.XHidden {
			newXHidden[k] = v
		}
		clone.XHidden = newXHidden
	}

	if len(e.Fields) > 0 {
		newFields := map[string]string{}
		for k, v := range e.Fields {
			newFields[k] = v
		}
		clone.Fields = newFields
	}
	return clone
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
func ErrorFromString(err string) *AError {
	if err == "" {
		return nil
	}

	err = strings.TrimPrefix(err, "#ERR ")
	e := &AError{}
	if er := json.Unmarshal([]byte(err), e); er != nil {
		return nil
	}
	return e
}

var ERRB = []byte("#ERR ")

// UnmarshalError unmarshal an error string to *Error
func UnmarshalError(err []byte) *AError {
	if len(err) == 0 {
		return nil
	}

	err = bytes.TrimPrefix(err, ERRB)
	e := &AError{}
	if er := json.Unmarshal(err, e); er != nil {
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
