package errormap

import "net/http"

// ErrorCode E0007

const (
	// 500
	ErrorCodeActiveVersionNotUnique = "E0007S0500N0001"

	// 404
	ErrorCodeActiveVersionNotFound = "E0007S0404N0001"
)

func init() {
	// 500
	register(ErrorCodeActiveVersionNotUnique, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，当前料号的启用版本不唯一，无法导入数据，请检查配置。",
		EN:    "Sorry, the active version of this material is not unique, please check the configuration.",
	})
	// 404
	register(ErrorCodeActiveVersionNotFound, http.StatusNotFound, langMap{
		ZH_CN: "对不起，未找到该料号的启用版本，无法导入数据，请检查配置。",
		EN:    "Sorry, the active version of this material is not found, please check the configuration.",
	})
}
