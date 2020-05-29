package errormap

import "net/http"

// ErrorCode E0001

const (
	// 400
	ErrorCodeAccountPasswordIncorrect = "E0000S0400N0001"
	// 500
	ErrorCodeLoginFailed = "E0000S0500N0001"
	// 401
	ErrorCodeUnauthenticated = "E0000S0401N0001"
)

func init() {
	register(ErrorCodeAccountPasswordIncorrect, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，账号或密码错误。",
		EN:    "Sorry, incorrect account or password.",
	})
	register(ErrorCodeLoginFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，登录失败，发生了一些错误。",
		EN:    "Sorry, cannot login with some internal error.",
	})
	register(ErrorCodeUnauthenticated, http.StatusForbidden, langMap{
		ZH_CN: "对不起，请先登录。",
		EN:    "Sorry, please login first.",
	})
}
