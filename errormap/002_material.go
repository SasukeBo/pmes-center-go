package errormap

import "net/http"

// ErrorCode E0002

const (
	// 400
	ErrorCodeMaterialAlreadyExists = "E0002S0400N0001"
	// 500
	ErrorCodeCreateSuccessButFetchFailed = "E0002S0500N0001"
)

func init() {
	register(ErrorCodeCreateSuccessButFetchFailed, http.StatusForbidden, langMap{
		ZH_CN: "抱歉，材料已创建，但由于某些错误未能获取数据，请向系统管理员寻求帮助。",
		EN:    "Sorry, material created but failed to fetch data with some errors, please ask help from your system administrator.",
	})
	register(ErrorCodeMaterialAlreadyExists, http.StatusForbidden, langMap{
		ZH_CN: "对不起，您创建的料号已经存在，请确认您的输入。",
		EN:    "Sorry, the material you created is already exists, please check your input.",
	})
}
