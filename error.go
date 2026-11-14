package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// special field
// + code
// + payload_string
type M map[string]any

type E string

func (e E) String() string {
	return string(e)
}

type ErrorAttribute struct {
	// Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"` // json encoded
	Type  string `json:"type,omitempty"`
}

type AError struct {
	cause error

	Id      string            `json:"id,omitempty"`
	Class   int32             `json:"class,omitempty"`  // remove http-code, should be derived from code
	Code    string            `json:"code,omitempty"`   // should be general database_error, access_deny
	Number  string            `json:"number,omitempty"` // unique, or hash of stack 4930543478 for grouping error
	Message map[string]string `json:"message,omitempty"`

	Attrs        map[string]*ErrorAttribute `json:"attrs,omitempty"`
	XHiddenAttrs map[string]*ErrorAttribute `json:"_hidden_attrs,omitempty" `

	TraceId string `json:"trace_id,omitempty" `
	SpanId  string `json:"span_id,omitempty" `
}

const E_none = ""
const E_invalid_input E = "invalid_input"
const E_invalid_url E = "invalid_url"
const E_unverified_domain E = "unverified_domain"
const E_unsupported_file_format E = "unsupported_file_format"
const E_api_http_unsupported E = "api_http_unsupported"
const E_meta_custom_audience_term_consent_required = "meta_custom_audience_term_consent_required"
const E_invalid_otp E = "invalid_otp"
const E_invalid_input_format E = "invalid_input_format"
const E_unassigned_number = "unassigned_number"
const E_invalid_field E = "invalid_field"
const E_email_taken E = "email_taken"
const E_invalid_domain E = "invalid_domain"
const E_missing_resource E = "missing_resource"
const E_inactive_resource E = "inactive_resource"
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
const E_invalid_kb_handle E = "invalid_kb_handle"
const E_kb_handle_already_used = "kb_handle_already_used"
const E_zalo_unsupported_version = "zalo_unsupported_version"
const E_zalo_account_locked = "zalo_account_locked"
const E_zalo_access_deny = "zalo_access_deny"
const E_zalo_zns_test_invalid_number = "zalo_zns_test_invalid_number"
const E_zalo_invalid_zns_template = "zalo_invalid_zns_template"
const E_zalo_zns_send_at_night = "zalo_zns_send_at_night"
const E_zalo_app_missing_permission = "zalo_app_missing_permission"
const E_zalo_zns_user_rejected = "zalo_zns_user_rejected"
const E_zalo_zns_oa_quota_exceeded = "zalo_zns_oa_quota_exceeded"
const E_zalo_zns_template_quota_exceeded = "zalo_zns_template_quota_exceeded"
const E_zalo_zns_template_manage_quota_exceeded = "zalo_zns_template_manage_quota_exceeded"
const E_timeout = "timeout"
const E_record_type_not_empty E = "record_type_not_empty"
const E_record_type_archived E = "record_type_archived"
const E_default_record_type_cannot_archive E = "default_record_type_cannot_archive"
const E_record_type_already_exists E = "record_type_already_exists"
const E_invalid_record_type_category E = "invalid_record_type_category"
const E_record_type_category_immutable E = "record_type_category_immutable"
const E_record_type_pipeline_immutable E = "record_type_pipeline_immutable"
const E_pipeline_already_exists E = "pipeline_already_exists"
const E_pipeline_in_use E = "pipeline_in_use"
const E_pipeline_stage_not_found E = "pipeline_stage_not_found"
const E_pipeline_stage_removal_not_supported E = "pipeline_stage_removal_not_supported"
const E_record_type_pipeline_missing E = "record_type_pipeline_missing"
const E_invalid_record_id E = "invalid_record_id"
const E_invalid_record_type_id E = "invalid_record_type_id"
const E_pipeline_stages_required E = "pipeline_stages_required"
const E_invalid_pipeline_stage_id E = "invalid_pipeline_stage_id"
const E_duplicate_pipeline_stage_id E = "duplicate_pipeline_stage_id"
const E_invalid_tag_id E = "invalid_tag_id"
const E_invalid_pipeline_stage_category E = "invalid_pipeline_stage_category"
const E_pipeline_stage_category_immutable E = "pipeline_stage_category_immutable"
const E_pipeline_stage_category_required E = "pipeline_stage_category_required"
const E_pipeline_update_in_progress E = "pipeline_update_in_progress"
const E_pipeline_stage_deleting E = "pipeline_stage_deleting"

// CDN domain errors, classified by the source of the failure.
const E_upload_idempotency_conflict E = "upload_idempotency_conflict"
const E_upload_size_mismatch E = "upload_size_mismatch"
const E_invalid_file_checksum E = "invalid_file_checksum"
const E_blocked_file_extension E = "blocked_file_extension"
const E_file_processing_failed E = "file_processing_failed"
const E_file_processing_lease_lost E = "file_processing_lease_lost"
const E_file_source_required E = "file_source_required"
const E_file_type_mismatch E = "file_type_mismatch"
const E_data_uri_requires_image E = "data_uri_requires_image"
const E_invalid_upload_metadata E = "invalid_upload_metadata"
const E_import_source_http_error E = "import_source_http_error"
const E_import_source_timeout E = "import_source_timeout"
const E_upload_session_expired E = "upload_session_expired"
const E_upload_session_closed E = "upload_session_closed"

func init() {
	ErrorTable[E_upload_session_expired] = H{"en_US": "The upload session has expired.", "vi_VN": "Phiên tải lên đã hết hạn."}
	ErrorTable[E_upload_session_closed] = H{"en_US": "The upload session is closed.", "vi_VN": "Phiên tải lên đã đóng."}
}

// copyErrorFields preserves the first-map convention without mutating caller data.
func copyErrorFields(fields ...M) M {
	field := M{}
	if len(fields) > 0 {
		for key, value := range fields[0] {
			field[key] = value
		}
	}
	return field
}

func EUploadSessionExpired(uploadID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["upload_id"] = uploadID
	return Error3("", nil, field, E_upload_session_expired, E_invalid_input)
}

func EUploadSessionClosed(uploadID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["upload_id"] = uploadID
	return Error3("", nil, field, E_upload_session_closed, E_invalid_input)
}

func ERecordTypeNotEmpty(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	return Error3(accid, nil, field, E_record_type_not_empty, E_invalid_input)
}

func ERecordTypeArchived(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	field["type"] = "record type"
	field["id"] = recordTypeID
	return Error3(accid, nil, field, E_record_type_archived, E_inactive_resource)
}

func EDefaultRecordTypeCannotArchive(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	return Error3(accid, nil, field, E_default_record_type_cannot_archive, E_invalid_input)
}

func ERecordTypeAlreadyExists(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	return Error3(accid, nil, field, E_record_type_already_exists, E_invalid_input)
}

func EInvalidRecordTypeCategory(accid, category string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["category"] = category
	field["name"] = "record_type.category"
	field["value"] = category
	return Error3(accid, nil, field, E_invalid_record_type_category, E_invalid_field, E_invalid_input)
}

func ERecordTypeCategoryImmutable(accid, recordTypeID, currentCategory, requestedCategory string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	field["current_category"] = currentCategory
	field["requested_category"] = requestedCategory
	return Error3(accid, nil, field, E_record_type_category_immutable, E_invalid_input)
}

func ERecordTypePipelineImmutable(accid, recordTypeID, currentPipelineID, requestedPipelineID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	field["current_pipeline_id"] = currentPipelineID
	field["requested_pipeline_id"] = requestedPipelineID
	return Error3(accid, nil, field, E_record_type_pipeline_immutable, E_invalid_input)
}

func EPipelineAlreadyExists(accid, pipelineID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	return Error3(accid, nil, field, E_pipeline_already_exists, E_invalid_input)
}

func EPipelineInUse(accid, pipelineID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	return Error3(accid, nil, field, E_pipeline_in_use, E_invalid_input)
}

func EPipelineStageNotFound(accid, pipelineID, pipelineStageID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	field["type"] = "pipeline stage"
	field["id"] = pipelineStageID
	return Error3(accid, nil, field, E_pipeline_stage_not_found, E_missing_resource, E_invalid_input)
}

func EPipelineStageRemovalNotSupported(accid, pipelineID, pipelineStageID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	return Error3(accid, nil, field, E_pipeline_stage_removal_not_supported, E_invalid_input)
}

func ERecordTypePipelineMissing(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	return Error3(accid, nil, field, E_record_type_pipeline_missing, E_internal)
}

func EInvalidRecordID(accid, recordID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_id"] = recordID
	field["name"] = "record.id"
	field["value"] = recordID
	return Error3(accid, nil, field, E_invalid_record_id, E_invalid_field, E_invalid_input)
}

func EInvalidRecordTypeID(accid, recordTypeID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["record_type_id"] = recordTypeID
	field["name"] = "record_type_id"
	field["value"] = recordTypeID
	return Error3(accid, nil, field, E_invalid_record_type_id, E_invalid_field, E_invalid_input)
}

func EPipelineStagesRequired(accid, pipelineID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	return Error3(accid, nil, field, E_pipeline_stages_required, E_invalid_input)
}

func EInvalidPipelineStageID(accid, pipelineID, pipelineStageID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	field["name"] = "pipeline_stage.id"
	field["value"] = pipelineStageID
	return Error3(accid, nil, field, E_invalid_pipeline_stage_id, E_invalid_field, E_invalid_input)
}

func EDuplicatePipelineStageID(accid, pipelineID, pipelineStageID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	return Error3(accid, nil, field, E_duplicate_pipeline_stage_id, E_invalid_input)
}

func EInvalidTagID(accid, tagID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["tag_id"] = tagID
	field["name"] = "tag_id"
	field["value"] = tagID
	return Error3(accid, nil, field, E_invalid_tag_id, E_invalid_field, E_invalid_input)
}

func EInvalidPipelineStageCategory(accid, pipelineID, pipelineStageID, category string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	field["category"] = category
	field["name"] = "pipeline_stage.category"
	field["value"] = category
	return Error3(accid, nil, field, E_invalid_pipeline_stage_category, E_invalid_field, E_invalid_input)
}

func EPipelineStageCategoryImmutable(accid, pipelineID, pipelineStageID, currentCategory, requestedCategory string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	field["current_category"] = currentCategory
	field["requested_category"] = requestedCategory
	return Error3(accid, nil, field, E_pipeline_stage_category_immutable, E_invalid_input)
}

func EPipelineStageCategoryRequired(accid, pipelineID, category string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["category"] = category
	return Error3(accid, nil, field, E_pipeline_stage_category_required, E_invalid_input)
}

func EPipelineUpdateInProgress(accid, pipelineID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	return Error3(accid, nil, field, E_pipeline_update_in_progress, E_invalid_input)
}

func EPipelineStageDeleting(accid, pipelineID, pipelineStageID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["pipeline_id"] = pipelineID
	field["pipeline_stage_id"] = pipelineStageID
	return Error3(accid, nil, field, E_pipeline_stage_deleting, E_invalid_input)
}

func EInvalidUrl(accid, url string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["url"] = url
	return Error3(accid, nil, field, E_invalid_url, E_invalid_input)
}

func EUnverifiedDomain(accid, domain string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["domain"] = domain
	return Error3(accid, nil, field, E_unverified_domain, E_invalid_domain, E_invalid_url, E_invalid_input)
}

func EInvalidDomain(accid, domain string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["domain"] = domain
	return Error3(accid, nil, field, E_invalid_domain, E_invalid_url, E_invalid_input)
}

func EUnsupportedFileFormat(accid, fileformat string, supporteds []string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_format"] = fileformat
	field["supported"] = strings.Join(supporteds, ", ")
	return Error3(accid, nil, field, E_unsupported_file_format, E_invalid_input)
}

func EApiHttpUnsupported(ip, url, ref string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["ip"] = ip
	field["_url"] = url
	field["_ref"] = ref
	return Error3("", nil, field, E_api_http_unsupported, E_invalid_input)
}

func EMetaCustomAudienceTermConsentRequired(adaccid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["ad_account_id"] = adaccid
	return Error3("", nil, field, E_meta_custom_audience_term_consent_required, E_invalid_input)
}

func EInvalidKbHandle(accid, kb, handle string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["kb"] = kb
	field["handle"] = handle
	return Error3(accid, nil, field, E_invalid_kb_handle, E_invalid_input)
}

func EDupKbHandle(accid, kb, handle string, fields ...M) *AError {
	field := copyErrorFields(fields...)

	field["kb"] = kb
	field["handle"] = handle
	return Error3(accid, nil, field, E_kb_handle_already_used, E_invalid_input)
}

func EClosePublicChannel(accid, convoid, channel string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["conversation_id"] = convoid
	field["channel"] = channel
	return Error3(accid, nil, field, E_close_public_channel, E_invalid_input)
}

func EPasswordTooWeak(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_password_too_weak, E_invalid_input)
}

func EGoogle(err error, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_google_error, E_internal, E_retryable)
}

func EInvalidPromotionCode(code string, errors []string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["promotion_code"] = code
	field["errors"] = errors
	return Error3("", nil, field, E_invalid_promotion_code, E_invalid_input)
}

func EInvalidCurrency(accid, cur string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["currency"] = cur
	return Error3(accid, nil, field, E_invalid_currency, E_invalid_input)
}

func EStillHaveOpenInvoice(accid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3(accid, nil, field, E_still_have_open_invoice, E_invalid_subscription, E_invalid_input)
}

func ELeaverIsTheLastOneInConvo(accid, convoid, issuerid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["conversation_id"] = convoid
	field["issuer_id"] = issuerid
	return Error3(accid, nil, field, E_leaver_is_the_last_one_in_conversation, E_access_deny, E_invalid_input)
}

func ENotAMember(accid, convoid, issuerid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["conversation_id"] = convoid
	field["issuer_id"] = issuerid
	return Error3(accid, nil, field, E_not_a_conversation_member, E_access_deny, E_invalid_input)
}

func EInvalidWebhookUrl(accid, webhookurl string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["webhook_url"] = webhookurl
	return Error3(accid, nil, field, E_invalid_webhook_url, E_invalid_input)
}

func EInvalidOTP(username, otp string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["username"] = username
	field["otp"] = otp
	return Error3("", nil, field, E_invalid_otp, E_invalid_input)
}

func EInactiveNumber(accid, number string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["number"] = number
	return Error3(accid, nil, field, E_inactive_number, E_invalid_integration, E_invalid_input)
}

func EBlockedNumber(accid, number string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["number"] = number
	return Error3(accid, nil, field, E_blocked_number, E_invalid_integration, E_invalid_input)
}

func EInviteLinkExpired(link string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["link"] = link
	return Error3("", nil, field, E_invite_link_expired, E_invalid_input)
}

func EInvalidIntegration(accid, inteid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["integration_id"] = inteid
	return Error3(accid, nil, field, E_invalid_integration, E_invalid_input)
}

func EMalformedRequest(accid, code string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["code"] = code
	return Error3(accid, nil, field, E_malformed_request, E_invalid_input)
}

func ENumberNotBounded(ctx context.Context, accid, number string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["number"] = number
	return Error3(accid, nil, field, E_unassigned_number, E_invalid_input)
}

func EConversationEnded(accid, convoid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["conversation_id"] = convoid
	return Error3(accid, nil, field, E_conversation_ended, E_invalid_input)
}

func EInvalidPasswordLength(currentLength, requiredLength int, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["current_length"] = currentLength
	field["required_length"] = requiredLength
	return Error3("", nil, field, E_invalid_password_length, E_invalid_input)
}

func EEmailTaken(email string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["email"] = email
	return Error3("", nil, field, E_email_taken, E_invalid_input)
}

func EInvalidPollConnection(accid, id string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["connection_id"] = id
	return Error3(accid, nil, field, E_invalid_connection, E_invalid_input)
}

func EDeadPollConnection(accid, id string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["connection_id"] = id
	return Error3(accid, nil, field, E_invalid_connection, E_invalid_input)
}

func EInvalidField(accid, name, value string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["name"] = name
	field["value"] = value
	return Error3(accid, nil, field, E_invalid_field, E_invalid_input)
}

func EDupplicateContactUpdate(accid, emailorphone string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["prop"] = emailorphone
	return Error3(accid, nil, field, E_dupplicate_contact_update, E_invalid_input)
}

func ENotEnoughCredit(accid, creditid, creditname, service string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["credit_id"] = creditid
	field["credit_name"] = creditname
	field["service"] = service
	return Error3(accid, nil, field, E_insufficient_credit)
}

func EInvalidZaloToken(accid, oaid, oaName string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["oa_id"] = oaid
	field["oa_name"] = oaName
	return Error3(accid, nil, field, E_invalid_zalo_token, E_service_unavailable, E_internal)
}

func EInvalidFacebookToken(accid, pageid, pageName string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["page_id"] = pageid
	field["page_name"] = pageName
	return Error3(accid, nil, field, E_invalid_facebook_token, E_service_unavailable, E_internal)
}

func EInvalidGoogleToken(accid, locationId, locationName string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["location_id"] = locationId
	field["location_name"] = locationName
	return Error3(accid, nil, field, E_invalid_google_token, E_service_unavailable, E_internal)
}

func EServiceUnavailable(err error, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", err, field, E_service_unavailable, E_internal, E_retryable)
}

func EPayloadTooLarge(curSize int64, maxSize int64, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["current_size"] = curSize
	field["maximum_size"] = maxSize
	return Error3("", nil, field, E_payload_too_large, E_invalid_input)
}

func ELimitExceeded(capacity int64, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["capacity"] = capacity
	return Error3("", nil, field, E_limit_exceeded, E_invalid_input)
}

func ERateLimit(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_rate_limit, E_limit_exceeded, E_invalid_input)
}

func EMissingId(typ string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_missing_id, E_invalid_input)
}

func EInvalidInputFormat(base error, fieldname, currentvalue string, msg string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["invalid_field"] = fieldname
	field["invalid_value"] = currentvalue
	field["msg"] = msg
	return Error3("", base, field, E_invalid_input_format, E_invalid_input)
}

func EInvalidInput(base error, required_fields []string, internal_message string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["required_fields"] = required_fields
	return Error3("", base, field, E_invalid_input)
}

func EInput(base error, f string, message string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["field"] = f
	field["message"] = message
	return Error3("", base, field, E_invalid_input)
}

func EAgentLocked(accid, agentid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["agent_id"] = agentid
	return Error3(accid, nil, field, E_locked_agent)
}

func EAccountLocked(accid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3(accid, nil, field, E_locked_account)
}

func ERetry(base error, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", base, field, E_internal, E_retryable)
}

func ETimeout(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_timeout, E_internal, E_retryable)
}

func EServer(base error, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", base, field, E_internal)
}

func EProvider(base error, external_service, action string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["external_service"] = external_service
	field["action"] = action
	return Error3("", base, field, E_provider_failed, E_internal)
}

func EParseInput(base error, payload []byte, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["size"] = len(payload)
	field["payload"] = Substring(string(payload), 0, 200)
	return Error3("", base, field, E_invalid_input)
}

func EData(base error, payload []byte, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["size"] = len(payload)
	field["payload"] = Substring(string(payload), 0, 200)
	return Error3("", base, field, E_internal, E_transform_data, E_data_corrupted)
}

func EInternalConnect(base error, fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", base, field, E_internal_connection, E_internal)
}

func EFS(base error, path string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["path"] = path
	return Error3("", base, field, E_file_system_error, E_internal)
}

func ELockedUser(userid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["user_id"] = userid
	return Error3("", nil, field, E_locked_user, E_access_deny)
}

func EDeny(userid string, requiredPerm string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["user_id"] = userid
	return Error3("", nil, field, E_access_deny)
}

func EExpiredAccessToken(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_access_deny, E_access_token_expired)
}

func EWrongPassword(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_wrong_password)
}

func EUnsub(accid, userid string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["user_id"] = userid
	return Error3(accid, nil, field, E_user_is_unsubscribed, E_invalid_input)
}

func EBanned(accid, userid string, internal_message string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["user_id"] = userid
	return Error3(accid, nil, field, E_user_is_banned)
}

func ErrContext(ctx context.Context, err error) error {
	return err
}

func EMissing(id, typ string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["type"] = typ
	field["id"] = id
	return Error3("", nil, field, E_missing_resource)
}

func ENotFound(id, typ string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["type"] = typ
	field["id"] = id
	return Error3("", nil, field, E_missing_resource)
}

func EInactive(id, typ string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["type"] = typ
	field["id"] = id
	return Error3("", nil, field, E_inactive_resource)
}

func EUploadIdempotencyConflict(uploadID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["upload_id"] = uploadID
	return Error3("", nil, field, E_upload_idempotency_conflict, E_invalid_input)
}

func EUploadSizeMismatch(expected, actual int64, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["expected_size"] = expected
	field["actual_size"] = actual
	return Error3("", nil, field, E_upload_size_mismatch, E_invalid_input)
}

func EInvalidFileChecksum(fileID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_id"] = fileID
	return Error3("", nil, field, E_invalid_file_checksum, E_internal)
}

func EBlockedFileExtension(name string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_name"] = name
	return Error3("", nil, field, E_blocked_file_extension, E_invalid_input)
}

func EFileProcessingFailed(fileID, causeCode string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_id"] = fileID
	field["cause_code"] = causeCode
	return Error3("", nil, field, E_file_processing_failed, E_internal)
}

func EFileProcessingLeaseLost(fileID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_id"] = fileID
	return Error3("", nil, field, E_file_processing_lease_lost, E_internal, E_retryable)
}

func EFileSourceRequired(fileID string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["file_id"] = fileID
	return Error3("", nil, field, E_file_source_required, E_invalid_input)
}

func EFileTypeMismatch(actual, requiredPrefix string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["mime_type"] = actual
	field["required_type_prefix"] = requiredPrefix
	return Error3("", nil, field, E_file_type_mismatch, E_invalid_input)
}

func EDataURIRequiresImage(mimeType string, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["mime_type"] = mimeType
	return Error3("", nil, field, E_data_uri_requires_image, E_invalid_input)
}

func EInvalidUploadMetadata(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_invalid_upload_metadata, E_internal)
}

func EImportSourceHTTPError(httpStatus int, fields ...M) *AError {
	field := copyErrorFields(fields...)
	field["source_http_status"] = httpStatus
	return Error3("", nil, field, E_import_source_http_error, E_invalid_input)
}

func EImportSourceTimeout(fields ...M) *AError {
	field := copyErrorFields(fields...)
	return Error3("", nil, field, E_import_source_timeout, E_invalid_input)
}

func IsErr(err error, code string) bool {
	var myerr *AError
	if !errors.As(err, &myerr) || myerr == nil {
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
	return newError(nil, err, field, codes, 1)
}

// newError owns construction and wrapping for both public constructors.
// stackSkip compensates for this helper and each constructor's caller convention.
func newError(ctx context.Context, err error, field M, codes []E, stackSkip int) *AError {
	if err != nil {
		var mye *AError
		if !errors.As(err, &mye) {
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
			mye.cause = err
			for key, value := range field {
				if key == "" {
					continue
				}
				valtype := getValueType(value)
				if key[0] == '_' {
					if mye.XHiddenAttrs == nil {
						mye.XHiddenAttrs = map[string]*ErrorAttribute{}
					}

					mye.XHiddenAttrs[key[1:]] = &ErrorAttribute{Value: rawify(value), Type: valtype}
				} else {
					if mye.Attrs == nil {
						mye.Attrs = map[string]*ErrorAttribute{}
					}
					mye.Attrs[key] = &ErrorAttribute{Value: rawify(value), Type: valtype}
				}
			}

			mye.Code = mergeErrorCodes(mye.Code, codes)
			if IsErr(mye, string(E_internal)) {
				mye.Class = 500
			}
			// Keep the originating error's grouping hash, but reflect retryability.
			if IsErr(mye, string(E_retryable)) && strings.HasPrefix(mye.Number, "SBQ-") {
				mye.Number = "SBR-" + strings.TrimPrefix(mye.Number, "SBQ-")
			}
			return mye
		}
	}

	outerr := &AError{Id: toBase58(time.Now().Unix()/3600) + toBase58(rand.Int64()), cause: err}
	// backward compatible, remove in future
	outerr.Class = 400
	retryable := false
	for _, code := range codes {
		if code == E_retryable {
			retryable = true
		}
		if code == E_internal {
			outerr.Class = 500
		}
	}
	outerr.Code = mergeErrorCodes("", codes)
	outerr.Attrs = map[string]*ErrorAttribute{}
	outerr.XHiddenAttrs = map[string]*ErrorAttribute{}

	for key, value := range field {
		if key == "" {
			continue
		}
		valtype := getValueType(value)
		if key[0] == '_' {
			outerr.XHiddenAttrs[key[1:]] = &ErrorAttribute{Value: rawify(value), Type: valtype}
		} else {
			outerr.Attrs[key] = &ErrorAttribute{Value: rawify(value), Type: valtype}
		}
	}

	if err != nil {
		outerr.XHiddenAttrs["root"] = &ErrorAttribute{Value: err.Error(), Type: "string"}
	}

	stack, funcname, funcstack := GetStack(stackSkip)
	outerr.Message = map[string]string{}
	if len(codes) > 0 {
		msg, has := ErrorTable[codes[0]]
		if has {
			outerr.Message["en_US"] = formatString(msg["en_US"], field)
			outerr.Message["vi_VN"] = formatString(msg["vi_VN"], field)
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

	// classified: database error, filesystem error, account access deny
	// SBR-72BFD5F
	// SBQ-A2BFD5F
	prefix := "Q"
	if retryable {
		prefix = "R"
	}
	if funcname != "" {
		outerr.XHiddenAttrs["function_name"] = &ErrorAttribute{Value: funcname, Type: "string"}
	}

	if field["_function_name"] != nil {
		if funcname, _ := field["_function_name"].(string); funcname != "" {
			outerr.XHiddenAttrs["function_name"] = &ErrorAttribute{Value: funcname, Type: "string"}
		}
	}

	errid := strings.ToUpper(strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(hostname+"/"+funcstack+"/"+outerr.Code))), 16))
	outerr.Number = "SB" + prefix + "-" + errid
	outerr.XHiddenAttrs["stack"] = &ErrorAttribute{Value: stack, Type: "string"}
	outerr.XHiddenAttrs["server_name"] = &ErrorAttribute{Value: hostname, Type: "string"}
	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		outerr.TraceId = span.SpanContext().TraceID().String()
		outerr.SpanId = span.SpanContext().SpanID().String()
	}
	return outerr
}

// Error creates an error and report it to server
// If you dont want to report it, use NewError instead
func Error(ctx context.Context, err error, code string, args ...any) *AError {
	if ctx == nil {
		ctx = context.Background()
	}
	codestrs := strings.Split(code, ",")
	codes := []E{}
	for _, codestr := range codestrs {
		codes = append(codes, E(strings.TrimSpace(codestr)))
	}
	outerr := NewError2(ctx, err, codes, args...)

	if span := trace.SpanFromContext(ctx); span != nil {
		span.RecordError(outerr)
		span.SetStatus(otelcodes.Error, outerr.Error())
	}

	args = append(args, "log_type")
	args = append(args, "error")

	b, _ := json.Marshal(outerr)

	emit(ctx, slog.LevelError, string(b), args...)
	return outerr
}

// NewError2 creates or wraps an error using context and key/value attributes.
func NewError2(ctx context.Context, err error, codes []E, args ...any) *AError {
	field := M{}
	for i := 0; i+1 < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok || key == "" || args[i+1] == nil {
			continue
		}
		field[key] = args[i+1]
	}
	return newError(ctx, err, field, codes, 2)
}

// Error creates an error and report it to server
// If you dont want to report it, use NewError instead
func Error3(accid string, err error, field M, codes ...E) *AError {
	if field != nil && accid == "" {
		if acci, has := field["account_id"]; has {
			accid, _ = acci.(string)
		}
	}

	if field != nil && accid != "" {
		field = copyErrorFields(field)
		field["account_id"] = accid
	}

	outerr := NewError(err, field, codes...)
	b, _ := json.Marshal(outerr)
	w.writeAndRetry(accid, LOG_ERR, string(b))
	return outerr
}

func (ae *AError) ToJSON() string {
	if ae == nil {
		return "null"
	}

	messageb, _ := json.Marshal(ae.Message)
	fieldb, _ := json.Marshal(ae.Attrs)
	out := `{"id": "` + ae.Id + `","code":` + fmt.Sprintf("%q", ae.Code) + `,"number":` + fmt.Sprintf("%q", ae.Number) + `,"attrs":` + string(fieldb) + `,"message":` + string(messageb) + ``
	return `{"code":` + fmt.Sprintf("%q", ae.Code) + `,"class":` + strconv.FormatInt(int64(ae.Class), 10) + `,"error":` + out + `}}`
}

func OverrideErrorTable(errtable map[E]H) {
	ErrorTable = errtable
}

func (e *AError) Clone() *AError {
	if e == nil {
		return nil
	}

	clone := &AError{
		cause:   e.cause,
		Id:      e.Id,
		Class:   e.Class,
		Code:    e.Code,
		Number:  e.Number,
		TraceId: e.TraceId,
		SpanId:  e.SpanId,
		Message: e.Message,
	}
	if len(e.XHiddenAttrs) > 0 {
		newXHiddenAttrs := map[string]*ErrorAttribute{}
		for k, v := range e.XHiddenAttrs {
			newXHiddenAttrs[k] = v
		}
		clone.XHiddenAttrs = newXHiddenAttrs
	}

	if len(e.Attrs) > 0 {
		newAttrs := map[string]*ErrorAttribute{}
		for k, v := range e.Attrs {
			newAttrs[k] = v
		}
		clone.Attrs = newAttrs
	}
	return clone
}

// Unwrap preserves the in-process error chain for errors.Is and errors.As.
// The cause is not serialized; decoded errors retain only their JSON metadata.
func (e *AError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func mergeErrorCodes(existing string, extra []E) string {
	codes := make([]string, 0, len(extra)+1)
	seen := map[string]bool{}
	appendCode := func(code string) {
		if code != "" && !seen[code] {
			seen[code] = true
			codes = append(codes, code)
		}
	}
	for _, code := range strings.Split(existing, ",") {
		appendCode(code)
	}
	for _, code := range extra {
		appendCode(string(code))
	}
	return strings.Join(codes, ",")
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

func getValueType(v any) string {
	switch v.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "number"
	case time.Duration:
		return "number" // ns
	case time.Time:
		return "time"
	case []int, []int64, []uint64, []float32, []float64, []string, []bool:
		return "array"
	case []byte:
		return "bytes"
	}
	return "object"
}

func rawify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// remove 0, I, O, l because its too hard to identify
const base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func toBase58(n int64) string {
	if n == 0 {
		return "1"
	}

	// Handle negative numbers
	if n < 0 {
		return toBase58(-n)
	}

	result := ""
	for n > 0 {
		result = string(base58Chars[n%58]) + result
		n /= 58
	}
	return result
}
