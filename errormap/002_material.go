package errormap

import "net/http"

// ErrorCode E0002

const (
	// 400
	ErrorCodeMaterialAlreadyExists                = "E0002S0400N0001"
	ErrorCodeMaterialAnalyzeMissingAttributeXAxis = "E0002S0400N0002"
	ErrorCodeMaterialAnalyzeMissingAttributeGroup = "E0002S0400N0003"
	ErrorCodeMaterialAnalyzeIllegalInput          = "E0002S0400N0004"
	// 500
	ErrorCodeCreateSuccessButFetchFailed = "E0002S0500N0001"
	ErrorCodeMaterialAnalyzeError        = "E0002S0500N0002"
)

func init() {
	// 400
	register(ErrorCodeMaterialAnalyzeIllegalInput, http.StatusForbidden, langMap{
		ZH_CN: "对不起，您输入参数不合法，请检查您的输入。",
		EN:    "Sorry, your input is illegal, please check your input.",
	})
	register(ErrorCodeMaterialAlreadyExists, http.StatusForbidden, langMap{
		ZH_CN: "对不起，您创建的料号已经存在，请确认您的输入。",
		EN:    "Sorry, the material you created is already exists, please check your input.",
	})
	register(ErrorCodeMaterialAnalyzeMissingAttributeXAxis, http.StatusForbidden, langMap{
		ZH_CN: "对不起，分析料号时缺少X轴字段，请选择。",
		EN:    "Sorry, missing the X-axis attribute, please choose one.",
	})
	register(ErrorCodeMaterialAnalyzeMissingAttributeGroup, http.StatusForbidden, langMap{
		ZH_CN: "对不起，分析料号时缺少分组字段，请选择。",
		EN:    "Sorry, missing the group attribute, please choose one.",
	})

	// 500
	register(ErrorCodeCreateSuccessButFetchFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "抱歉，材料已创建，但由于某些错误未能获取数据，请向系统管理员寻求帮助。",
		EN:    "Sorry, material created but failed to fetch data with some errors, please ask help from your system administrator.",
	})
	register(ErrorCodeMaterialAnalyzeError, http.StatusInternalServerError, langMap{
		ZH_CN: "抱歉，分析料号数据时发生了一些错误，分析失败。",
		EN:    "Sorry, failed to analyze the data of the material with some errors.",
	})
}
