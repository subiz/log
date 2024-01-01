
/* GENERATED FILE, DO NOT EDIT */
package log

type H map[string]string

var ErrorTable = map[E]H{
	"invalid_integration": H{
		"vi_VN": "Cài đặt kênh giao tiếp bị ngắt hoặc không tồn tại. Vui lòng cài dặt lại kênh giao tiếp",
		"en_US": "Communication channel settings are disconnected or do not exist. Please reset the communication channel",
	},
	"invalid_password_length": H{
		"vi_VN": "Mật khẩu quá ngắn, vui lòng chọn mật khẩu nhiều hơn {required_length} ký tự",
		"en_US": "Password is too short, please choose a password with more than {required_length} characters",
	},
	"conversation_ended": H{
		"vi_VN": "Hội thoại đã kết thúc, bạn không thể thực hiện hành động trọng hội thoại này",
		"en_US": "The conversation has ended, you cannot perform actions in this conversation",
	},
	"invalid_token": H{
		"vi_VN": "Mã không hợp lệ",
		"en_US": "The token is not valid",
	},
	"malformed_request": H{
		"vi_VN": "Yêu cầu không đúng. Vui lòng liên hệ Subiz để được hỗ trợ",
		"en_US": "Malformed request. Please contact Subiz for support",
	},
	"email_taken": H{
		"vi_VN": "Email {email} đã được sử dụng, vui lòng sử dụng email khác",
		"en_US": "Email {email} is already taken, please use another email",
	},
	"invalid_connection": H{
		"vi_VN": "Kết nối không hợp lệ, vui lòng kết nối lại",
		"en_US": "Your connection is invalid, please reconnect",
	},
	"invalid_poll_connection": H{
		"vi_VN": "Kết nối thời gian thực không hợp lệ, vui lòng kết nối lại",
		"en_US": "Real-time connection is invalid, please reconnect",
	},
	"dead_poll_connection": H{
		"vi_VN": "Kết nối thời gian thực đã bị ngắt do nghẽn mạng, vui lòng thử lại sau",
		"en_US": "Real-time connection was interrupted due to network congestion, please try again later",
	},
	"duplicate_contact": H{
		"vi_VN": "Giá trị {prop} đã được sử dụng cho một hồ sơ khách khác. Để có thể cập nhật, bạn cần gộp 2 hồ sơ hoặc xóa giá trị này ở hồ sơ còn lại.",
		"en_US": "The property {prop} is already in used for another contact. To continue, please merge 2 contacts or remove this value in the other contact.",
	},
	"insufficient_credit": H{
		"vi_VN": "Tài khoản của bạn không còn đủ credit. Vui lòng thanh toán tài khoản '{credit_name}' để tiếp tục",
		"en_US": "Your account has insufficient credits. Please pay for billing account '{credit_name}' to continue",
	},
	"invalid_facebook_token": H{
		"vi_VN": "Kết nối tới Facebook đã hết hạn. Vui lòng tích hợp lại fanpage {page_name} ({page_id}) để tiếp tục",
		"en_US": "Connection to Facebook has expired. Please reintegrate fanpage {page_name} ({page_id}) to continue",
	},
	"invalid_field": H{
		"vi_VN": "Trường dữ liệu {name} không hợp lệ. Vui lòng thử lại với giá trị khác",
		"en_US": "The field {name} is not valid. Please try again with a different value",
	},
	"invalid_zalo_token": H{
		"vi_VN": "Kết nối tới Zalo đã hết hạn. Vui lòng tích hợp lại OA {oa_name} ({oa_id}) để tiếp tục",
		"en_US": "Connection to Zalo has expired. Please reintegrate OA {oa_name} ({oa_id}) to continue",
	},
	"invalid_google_token": H{
		"vi_VN": "Kết nối tới tài khoản Google đã hết hạn. Vui lòng tích hợp lại địa điểm kinh doanh {location_name} ({location_id}) để tiếp tục",
		"en_US": "Connection to Google Account has expired. Please reintegrate your business location {location_name} ({location_id}) to continue",
	},
	"service_unavailable": H{
		"vi_VN": "Không thể kết nối tới dịch vụ cần thiết. Vui lòng thử lại sau",
		"en_US": "Unable to connect to the required service. Please try again later",
	},
	"payload_too_large": H{
		"vi_VN": "Dung lượng gói tin quá lớn",
		"en_US": "Payload too large",
	},
	"limit_exceeded": H{
		"vi_VN": "Bạn đã sử dụng quá giới hạn cho phép",
		"en_US": "You have exceeded the allowable limit",
	},
	"invalid_domain": H{
		"vi_VN": "",
		"en_US": "Something wrong, please try again later",
	},
	"missing_id": H{
		"vi_VN": "Lỗi định danh {type} không hợp lệ. Vui lòng cung cấp đầy đủ định danh hoặc liên hệ Subiz để được hỗ trợ",
		"en_US": "Invalid identify for {type}. Please provide the corrected identify or contact Subiz for support",
	},
	"not_a_conversation_member": H{
		"vi_VN": "Bạn không phải là thành viên của hội thoại. Bạn cần được mời vào hội thoại để tiếp tục",
		"en_US": "You are not a member of this conversation. You need to be invited before continue this action",
	},
	"transform_data": H{
		"vi_VN": "",
		"en_US": "Something wrong, please try again later",
	},
	"provider_failed": H{
		"vi_VN": "Yêu cầu thất bại từ {external_service}. Vui lòng thử lại sau",
		"en_US": "Your request to {external_service} failed. Please try again later",
	},
	"locked_account": H{
		"vi_VN": "Tài khoản của bạn đang bị khóa. Vui lòng liên hệ chủ tài khoản hoặc Subiz để được hỗ trợ",
		"en_US": "Your account is locked. Please contact account owner or Subiz for support",
	},
	"locked_agent": H{
		"vi_VN": "Tài khoản của bạn đang bị khóa. Vui lòng liên hệ chủ tài khoản hoặc Subiz để được hỗ trợ",
		"en_US": "Your account is locked. Please contact account owner or Subiz for support",
	},
	"provider_data_mismatched": H{
		"vi_VN": "Bất đồng bộ dữ liệu với {external_service}.",
		"en_US": "Data type mismatch with {external_service}.",
	},
	"file_system_error": H{
		"vi_VN": "Lỗi hệ thống tệp. Vui lòng thử lại hoặc liên hệ Subiz để được hỗ trợ",
		"en_US": "File system error. Please retry or contact Subiz for support",
	},
	"access_token_expired": H{
		"vi_VN": "Mã truy cập đã hết hạn. Vui lòng đăng nhập lại hoặc xin mã mới",
		"en_US": "Access token is expired. Please login again or request a new access token",
	},
	"internal_connection": H{
		"vi_VN": "Lỗi kết nối nội bộ, vui lòng thử lại sau",
		"en_US": "Internal connection error, please retry later",
	},
	"missing_resource": H{
		"vi_VN": "Không tìm thấy {type}",
		"en_US": "{type} not found",
	},
	"not_found": H{
		"vi_VN": "Không tìm thấy {type}",
		"en_US": "{type} not found",
	},
	"internal": H{
		"vi_VN": "Lỗi hệ thống. Vui lòng thử lại sau",
		"en_US": "System error. Please try again later",
	},
	"access_deny": H{
		"vi_VN": "Từ chối truy cập",
		"en_US": "Access deny",
	},
	"locked_user": H{
		"vi_VN": "Tài khoản của bạn đã bị khóa",
		"en_US": "Your account had been locked",
	},
	"unauthorized": H{
		"vi_VN": "Bạn không có đủ quyền. Để thực hiện chức năng này bạn cần thêm quyền {mising}",
		"en_US": "You are not authorized. To perform this action you need following permission: {missing}",
	},
	"invalid_input": H{
		"vi_VN": "Dữ liệu đầu vào không hợp lệ",
		"en_US": "Invalid input data",
	},
	"invalid_email": H{
		"vi_VN": "Email không hợp lệ",
		"en_US": "Email is not valid",
	},
	"weak_password": H{
		"vi_VN": "Mật khẩu quá yếu. Mật khẩu phải chứa ít nhất 8 ký tự",
		"en_US": "Password is too week. Password should contains at least 8 chracters",
	},
	"user_is_banned": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
	"user_is_unsubscribed": H{
		"vi_VN": "người dùng này đã từ chối nhận tin truyền thông của bạn, vui lòng bỏ người dùng khỏi danh sách từ chối nhận để tiếp tục",
		"en_US": "This user have unsubscribed your marketing messages, please remove this usser from unsubscribe list to continue",
	},
	"agent_is_inactived": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
	"wrong_password": H{
		"vi_VN": "Sai mật khẩu",
		"en_US": "Wrong password",
	},
	"password_too_weak": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
}
