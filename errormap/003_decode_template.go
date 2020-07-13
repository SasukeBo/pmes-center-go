package errormap

import "net/http"

// ErrorCode E0003

const (
	// 400
	ErrorCodeDecodeTemplateSetDefaultFailed     = "E0003S0400N0001"
	ErrorCodeDecodeTemplateDefaultDeleteProtect = "E0003S0400N0002"
)

func init() {
	// 400
	register(ErrorCodeDecodeTemplateDefaultDeleteProtect, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，不能删除料号的默认解析模板。",
		EN:    "Sorry, the default decode template cannot be deleted.",
	})
	register(ErrorCodeDecodeTemplateSetDefaultFailed, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，设置料号默认模板时发生错误。",
		EN:    "Sorry, failed to set default decode template for the material with some errors.",
	})
}
