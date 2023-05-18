package log

type H map[string]string

var ErrorTable = map[E]H{
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
	"agent_is_inactived": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
	"wrong_password": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
	"password_too_weak": H{
		"vi_VN": "Người dùng này đã bị chặn, vui lòng bỏ chặn để tiếp tục",
		"en_US": "This user have been banned, please unban this user to continue",
	},
}
