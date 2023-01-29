package log

import (
	"context"
	"errors"
	"fmt"

	"github.com/subiz/header"
)

type M map[string]interface{}

type E int64

const E_missing_field E = 3
const E_invalid_field E = 4
const E_resource_not_found E = 6
const E_access_deny E = 7
const E_account_not_found E = 9
const E_invalid_credential E = 10
const E_http_call_error E = 11
const E_agent_group_not_found E = 12
const E_invalid_conversation_id E = 13
const E_agent_not_found E = 15
const E_facebook_page_not_found E = 16
const E_invalid_refresh_token E = 17
const E_wrong_password E = 18
const E_agent_is_not_active E = 19
const E_invalid_token E = 20
const E_invalid_message_size E = 21
const E_invalid_payload_size E = 22
const E_user_is_banned E = 26
const E_wrong_signature E = 27
const E_user_not_found E = 33
const E_invalid_attribute_value E = 35
const E_invalid_attribute_key E = 36
const E_invalid_attribute_name E = 37
const E_invalid_attribute_type E = 38
const E_too_many_attribute E = 39
const E_attribute_not_found E = 40
const E_invalid_agent_id E = 41
const E_invalid_note_target_id E = 42
const E_invalid_event_id E = 43
const E_invalid_content_type E = 44
const E_parse_query_failed E = 45
const E_domain_is_not_whitelisted E = 46
const E_invalid_domain E = 47
const E_ip_is_blocked E = 48
const E_invalid_transaction_id E = 49
const E_invalid_amount E = 50
const E_inactive_promotion_referral_program E = 51
const E_plan_not_found E = 52
const E_invoice_not_found E = 53
const E_invalid_stripe_info E = 54
const E_subscription_not_found E = 55
const E_invalid_plan E = 56
const E_promotion_not_found E = 57
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
const E_payment_method_not_found E = 79
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
const E_template_message_key_not_found E = 96
const E_template_message_not_is_creator E = 97
const E_product_failed E = 98
const E_product_not_found E = 99
const E_order_not_found E = 100
const E_invalid_base_currency E = 101
const E_invalid_currency E = 102
const E_invalid_url E = 103
const E_invalid_poll_token E = 104
const E_dead_poll_connection E = 105
const E_connection_not_found E = 106
const E_pos_not_found E = 107
const E_too_many_pos E = 108
const E_tax_not_found E = 109
const E_too_many_tax E = 110
const E_too_many_shipping_address E = 111
const E_shipping_address_not_found E = 112
const E_message_not_found E = 113
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
const E_conversation_not_found E = 133
const E_too_many_fields E = 134
const E_too_many_attachments E = 135
const E_unknown_message_format E = 136
const E_invalid_field_size E = 137
const E_shopee_call_error E = 138
const E_unauthorized_shopee_shop E = 139
const E_invalid_ratelimit_config E = 140
const E_too_many_requests E = 141
const E_order_readonly E = 142
const E_shipping_policy_not_found E = 143
const E_too_many_shipping_policy E = 144
const E_pipeline_stage_not_found E = 145
const E_invalid_order_status E = 146
const E_marker_not_found E = 147
const E_invalid_product_category E = 148
const E_invalid_product_visibility E = 149
const E_invalid_product_id E = 150
const E_shortcut_not_found E = 151
const E_file_system_error E = 152
const E_cannot_connect E = 153
const E_zalo_call_error E = 154
const E_invalid_integration_state E = 155
const E_message_too_large E = 156
const E_invalid_message_id E = 157
const E_attachment_too_large E = 158
const E_field_too_long E = 159
const E_request_timeout E = 160
const E_invalid_conversation_modal_secret E = 170
const E_conversation_modal_not_found E = 171
const E_invalid_conversation_modal_url E = 172
const E_invalid_conversation_modal_key E = 173
const E_conversation_automation_not_found E = 174
const E_sendgrid_error E = 176
const E_too_many_events E = 177
const E_too_many_number E = 178
const E_too_many_phone_device E = 179
const E_phone_device_not_found E = 180
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
const E_call_not_found E = 205
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
const E_database_error E = 5003
const E_json_data_corrupted E = 5008     // json payload is broken
const E_protobuf_data_corrupted E = 5004 // protobuf message is broken
const E_facebook_call_failed E = 5023
const E_invalid_facebook_access_token E = 5024
const E_subiz_call_failed E = 5025
const E_data_corrupted E = 5028
const E_invalid_google_auth_response E = 5071
const E_pdf_generate_failed E = 5072

func ErrDatabase(ctx context.Context, accid string, err error, query string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["query"] = query
	return NewError(ctx, err, E_database_error, internal_message, field)
}

func ErrJSON(ctx context.Context, accid string, err error, size int64, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["size"] = size
	return NewError(ctx, err, E_json_data_corrupted, internal_message, field)
}

func ErrProto(ctx context.Context, accid string, err error, size int64, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["size"] = size
	return NewError(ctx, err, E_protobuf_data_corrupted, internal_message, field)
}

func ErrFileSystem(ctx context.Context, accid string, err error, path string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["path"] = path
	return NewError(ctx, err, E_file_system_error, internal_message, field)
}

func ErrAccessDeny(ctx context.Context, accid, userid string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["user_id"] = userid
	return NewError(ctx, nil, E_user_is_banned, internal_message, field)
}

func ErrUserIsBanned(ctx context.Context, accid, userid string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["user_id"] = userid
	return NewError(ctx, nil, E_user_is_banned, internal_message, field)
}

func ErrNotFound(ctx context.Context, accid, typ, hint string, internal_message string, fields ...M) error {
	var field = M{}
	if len(fields) > 0 && fields[0] != nil {
		field = fields[0]
	}
	field["account_id"] = accid
	field["type"] = typ
	field["hint"] = hint
	return NewError(ctx, nil, E_resource_not_found, internal_message, field)
}

func NewError(ctx context.Context, err error, code E, internal_message string, field M) error {
	if err == nil {
		err = errors.New(internal_message)
	} else {
		err = fmt.Errorf("%w %s", err, internal_message)
	}
	if sentryDsn != "" {
		return NewSentryErr(ctx, err, code, internal_message, field)
	}
	return &header.Error{}
	// log host name
	// log ip
}
