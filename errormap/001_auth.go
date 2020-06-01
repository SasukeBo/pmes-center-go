package errormap

import "net/http"

// ErrorCode E0001

const (
	// 400
	ErrorCodeAccountPasswordIncorrect = "E0001S0400N0001"
	ErrorCodeAccountNotExistError     = "E0001S0400N0002"
	ErrorCodeLogoutFailedError        = "E0001S0400N0003"
	// 500
	ErrorCodeLoginFailed          = "E0001S0500N0001"
	ErrorCodeLogoutFailed         = "E0001S0500N0002"
	ErrorCodeRecordUserLoginError = "E0001S0500N0003"
	// 401
	ErrorCodeUnauthenticated      = "E0001S0401N0001"
	ErrorCodePasswordChangedError = "E0001S0401N0002"
	// 403
	ErrorCodePermissionDeny = "E0001S0403N0001"
)

func init() {
	register(ErrorCodePermissionDeny, http.StatusForbidden, langMap{
		ZH_CN: "对不起，您没有权限进行此操作。",
		EN:    "Sorry, you have no permission to do this operation.",
	})
	register(ErrorCodePasswordChangedError, http.StatusUnauthorized, langMap{
		ZH_CN: "对不起，用户密码已修改，请重新登录。",
		EN:    "Sorry, the password of this account has been changed, please login again.",
	})
	register(ErrorCodeLogoutFailedError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，退出登录时发生错误。",
		EN:    "Sorry, log out failed with some errors.",
	})
	register(ErrorCodeAccountNotExistError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，账号不存在。",
		EN:    "Sorry, the account not exists.",
	})
	register(ErrorCodeAccountPasswordIncorrect, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，账号或密码错误。",
		EN:    "Sorry, incorrect account or password.",
	})
	register(ErrorCodeLoginFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，登录失败，发生了一些错误。",
		EN:    "Sorry, cannot login with some internal errors.",
	})
	register(ErrorCodeRecordUserLoginError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，记录登录状态时发生错误。",
		EN:    "Sorry, record login failed with some errors.",
	})
	register(ErrorCodeLogoutFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，登出失败，发生了一些错误。",
		EN:    "Sorry, cannot log out with some internal error.",
	})
	register(ErrorCodeUnauthenticated, http.StatusForbidden, langMap{
		ZH_CN: "对不起，请先登录。",
		EN:    "Sorry, please login first.",
	})
}
