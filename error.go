package log

import (
	"context"
	"encoding/json"
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"github.com/subiz/header"
)

type M map[string]interface{}

type E string

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_not_found E = "not_found"
const E_access_deny E = "access_deny"
const E_internal E = "internal"

// //////////////// SUB CODE, for advanced branching ///////////////////////////
const E_database_error E = "database_error"
const E_file_system_error E = "file_system_error"
const E_transform_data E = "transform_data" // json payload is broken

const E_locked_user E = "locked_user"
const E_unauthorized E = "unauthorized"
const E_wrong_password E = "wrong_password"
const E_user_is_banned E = "user_is_banned"
const E_wrong_signature E = "wrong_signature"

/*
const E_invalid_credential E = 10
const E_http_call_error E = 11
const E_invalid_conversation_id E = 13
const E_invalid_refresh_token E = 17
const E_agent_is_not_active E = 19
const E_invalid_token E = 20
const E_invalid_message_size E = 21
const E_invalid_payload_size E = 22
const E_invalid_attribute_value E = 35
const E_invalid_attribute_key E = 36
const E_invalid_attribute_name E = 37
const E_invalid_attribute_type E = 38
const E_too_many_attribute E = 39
const E_invalid_agent_id E = 41
const E_invalid_note_target_id E = 42
const E_invalid_content_type E = 44
const E_parse_query_failed E = 45
const E_domain_is_not_whitelisted E = 46
const E_invalid_domain E = 47
const E_ip_is_blocked E = 48
const E_invalid_transaction_id E = 49
const E_invalid_amount E = 50
const E_inactive_promotion_referral_program E = 51
const E_invalid_stripe_info E = 54
const E_invalid_plan E = 56
const E_invalid_promotion E = 58
const E_invalid_email E = 59
const E_email_taken E = 60
const E_invalid_date_format E = 61
const E_prohibited_action E = 62
const E_invalid_working_day E = 63
const E_invalid_holiday E = 64
const E_invalid_fullname E = 66
const E_invalid_password E = 67
const E_invalid_token_type E = 68
const E_invalid_invoice_template E = 69
const E_invoice_template_compile_failed E = 70
const E_invalid_stripe_customer_id E = 73
const E_invalid_stripe_token E = 74
const E_stripe_call_failed E = 75
const E_invalid_bill_id E = 76
const E_invalid_invoice_id E = 77
const E_invalid_id E = 78
const E_invalid_invoice_duedate E = 80
const E_filestore_error E = 81
const E_invalid_country E = 82
const E_invalid_referrer_code E = 83
const E_invalid_oauth_scope E = 84
const E_invalid_label_id E = 85
const E_invalid_label_name E = 86
const E_invalid_label_color E = 87
const E_invalid_label_description E = 88
const E_invalid_user_view_name E = 89
const E_limit_reached E = 90
const E_invalid_time_range E = 91
const E_empty_file E = 92
const E_invalid_file E = 93
const E_endchat_bot_setting_after_any_message_too_low E = 94
const E_endchat_bot_setting_after_any_message_too_high E = 95
const E_template_message_not_is_creator E = 97
const E_product_failed E = 98
const E_invalid_base_currency E = 101
const E_invalid_currency E = 102
const E_invalid_url E = 103
const E_invalid_poll_token E = 104
const E_dead_poll_connection E = 105
const E_too_many_pos E = 108
const E_too_many_tax E = 110
const E_too_many_shipping_address E = 111
const E_expired_access_token E = 114
const E_too_many_payment_method E = 115
const E_invalid_css E = 116
const E_invalid_report_range E = 117
const E_tempfile_error E = 118
const E_invalid_goal_status E = 119
const E_user_is_not_in_the_conversation E = 120
const E_invalid_integration E = 121
const E_invalid_message_type E = 122
const E_missing_pong E = 123
const E_invalid_pong_type E = 124
const E_invalid_event_type E = 125
const E_remover_is_not_agent E = 127
const E_user_is_the_last_one_in_conversation E = 128
const E_conversation_ended E = 129
const E_leaver_is_the_last_one_in_conversation E = 130
const E_invalid_conversation_state E = 131
const E_caller_is_not_leaver E = 132
const E_too_many_fields E = 134
const E_too_many_attachments E = 135
const E_unknown_message_format E = 136
const E_invalid_field_size E = 137
const E_shopee_call_error E = 138
const E_unauthorized_shopee_shop E = 139
const E_invalid_ratelimit_config E = 140
const E_too_many_requests E = 141
const E_order_readonly E = 142
const E_too_many_shipping_policy E = 144
const E_invalid_order_status E = 146
const E_invalid_product_category E = 148
const E_invalid_product_visibility E = 149
const E_invalid_product_id E = 150
const E_cannot_connect E = 153
const E_zalo_call_error E = 154
const E_invalid_integration_state E = 155
const E_message_too_large E = 156
const E_invalid_message_id E = 157
const E_attachment_too_large E = 158
const E_field_too_long E = 159
const E_request_timeout E = 160
const E_invalid_conversation_modal_secret E = 170
const E_invalid_conversation_modal_url E = 172
const E_invalid_conversation_modal_key E = 173
const E_sendgrid_error E = 176
const E_too_many_events E = 177
const E_too_many_number E = 178
const E_too_many_phone_device E = 179
const E_sip_server_error E = 181
const E_invalid_phone_device_state E = 182
const E_invalid_call_state E = 190
const E_number_is_blocked E = 191
const E_number_is_not_active E = 192
const E_number_is_registered E = 193
const E_number_is_not_bounded E = 194
const E_too_many_greeting_audio E = 195
const E_file_too_large E = 196
const E_unsupported_file_type E = 197
const E_invalid_sip_provider E = 198 // fpt, itel
const E_too_many_agent E = 199
const E_duplicated_extension E = 200
const E_apple_login_error E = 201
const E_wrong_shard E = 202
const E_webrtc_failed E = 203
const E_out_of_port E = 204
const E_bad_webrtc_track E = 206
const E_invalid_webrtc_negotiation E = 207
const E_bad_webrtc_connection E = 208
const E_user_duplicated E = 209
const E_invalid_task_title E = 210
const E_invalid_task_note E = 211
const E_invalid_task_status E = 213
const E_invalid_task_assignee E = 214
const E_invalid_task_supervisor E = 215
const E_invalid_task_watchers E = 216
const E_undefined E = 5001 // unknown error
const E_not_implemented E = 5002
const E_protobuf_data_corrupted E = 5004 // protobuf message is broken
const E_facebook_call_failed E = 5023
const E_invalid_facebook_access_token E = 5024
const E_subiz_call_failed E = 5025
const E_data_corrupted E = 5028
const E_invalid_google_auth_response E = 5071
const E_pdf_generate_failed E = 5072
*/

func ErrInvalidInput(ctx context.Context, err error, required_fields []string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["required_fields"] = required_fields
	return NewError(ctx, err, field, E_invalid_input)
}

func ErrServer(ctx context.Context, err error, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return NewError(ctx, err, field, E_internal)
}

func ErrDB(ctx context.Context, err error, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return NewError(ctx, err, field, E_internal, E_database_error)
}

func ErrData(ctx context.Context, err error, payload []byte, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["size"] = len(payload)
	field["payload"] = string(payload[:200])
	return NewError(ctx, err, field, E_internal, E_transform_data)
}

func ErrFS(ctx context.Context, err error, path string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["path"] = path
	return NewError(ctx, err, field, E_internal, E_file_system_error)
}

func ErrLockedUser(ctx context.Context, userid string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return NewError(ctx, nil, field, E_locked_user, E_access_deny)
}

func ErrUnauthorized(ctx context.Context, userid string, missingPerms []string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	field["missing"] = strings.Join(missingPerms, ",")
	return NewError(ctx, nil, field, E_access_deny, E_unauthorized)
}

func ErrAccessDeny(ctx context.Context, userid string, requiredPerm string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["user_id"] = userid
	return NewError(ctx, nil, field, E_access_deny)
}

func ErrWrongPassword(ctx context.Context, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	return NewError(ctx, nil, field, E_wrong_password)
}

func ErrUserIsBanned(ctx context.Context, accid, userid string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["user_id"] = userid
	return NewError(ctx, nil, field, E_user_is_banned)
}

func ErrNotFound(ctx context.Context, id, typ string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["type"] = typ
	field["id"] = id
	return NewError(ctx, nil, field, E_not_found)
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

func NewError(ctx context.Context, err error, field M, codes ...E) error {
	/*
		if sentryDsn != "" {
			accid := ""
			if field != nil {
				accidi := field["account_id"]
				if accidi != nil {
					accid, _ = accidi.(string)
				}
			}
			return NewSentryErr(ctx, accid, err, code, internal_message, field)
		}
	*/

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

	stack := getStack(0)
	numberi, hasNumber := field["number"]
	var number int64
	if hasNumber {
		switch v := numberi.(type) {
		case int:
			number = int64(v)
		case int32:
			number = int64(v)
		case int64:
			number = int64(v)
		case uint32:
			number = int64(v)
		case uint64:
			number = int64(v)
		case string:
			number, _ = strconv.ParseInt(v, 10, 0)
		}
	}

	if len(codes) > 0 {
		msg, has := ErrorTable[codes[0]]
		if has {
			outerr.Message = &header.I18NString{
				En_US: formatString(msg["en_US"], field),
				Vi_VN: formatString(msg["vi_VN"], field),
			}
		}
	}

	// compute number based on stack trace
	if number == 0 {
		number = int64(crc32.ChecksumIEEE([]byte(stack)))
	}
	outerr.Number = number

	outerr.XHidden["stack"] = stack
	outerr.XHidden["server_name"] = hostname

	if serverEnv != "" {
		// hostname
		metricmaplock.Lock()
		metricmap[number] = &header.Event{AccountId: outerr.XHidden["account_id"], Created: time.Now().UnixMilli(), UserId: outerr.XHidden["user_id"], Data: &header.Data{Error: outerr}}
		metricmapcount[number]++
		metricmaplock.Unlock()
	}
	return outerr
}
