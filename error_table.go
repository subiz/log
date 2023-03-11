package log

type H map[string]string

var ErrorTable = map[E]H{
	"transform_data": H{
		"vi_VN": "",
		"en_US": "Something wrong, please try again later",
	},
	"file_system_error": H{
		"vi_VN": "Lỗi hệ thống tệp",
		"en_US": "File system error",
	},
	"database_error": H{
		"vi_VN": "",
		"en_US": "Something wrong, please try again later",
	},
	"1": H{
		"code":  "connection",
		"vi_VN": "",
		"en_US": "Something wrong, please try again later",
	},
	"not_found": H{
		"vi_VN": "Không tìm thấy {type}",
		"en_US": "{type} not found",
	},
	"internal": H{
		"vi_VN": "Lỗi hệ thống",
		"en_US": "System error",
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
