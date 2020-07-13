package errormap

import "net/http"

// ErrorCode E0005

const (
	// 400
	ErrorCodePointAlreadyExists = "E0005S0400N0001"
)

func init() {
	// 400
	register(ErrorCodePointAlreadyExists, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，您创建的检测项已经存在，请确认您的输入。",
		EN:    "Sorry, the detect item you created already exists, please check your input.",
	})
}
