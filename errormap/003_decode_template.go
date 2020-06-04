package errormap

import "net/http"

// ErrorCode E0003

const (
	// 400
	ErrorCodeDecodeTemplateSetDefaultFailed = "E0003S0400N0001"
)

func init() {
	register(ErrorCodeDecodeTemplateSetDefaultFailed, http.StatusForbidden, langMap{
		ZH_CN: "对不起，设置料号默认模板时发生错误。",
		EN:    "Sorry, failed to set default decode template for the material with some errors.",
	})
}
