package errormap

import "net/http"

// ErrorCode E0006

const (
	// 500
	ErrorCodeRevertImportFailed = "E0006S0500N0001"
	ErrorCodeAssembleDataFailed = "E0006S0500N0002"
	// 400
	ErrorCodeImportGetPointsFailed           = "E0006S0400N0001"
	ErrorCodeImportWithIllegalDecodeTemplate = "E0006S0400N0002"
	ErrorCodeImportFailedWithPanic           = "E0006S0400N0003"
)

func init() {
	// 400
	register(ErrorCodeImportFailedWithPanic, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，导入数据时发生了未知错误。",
		EN:    "Sorry, failed to import data with some unknown errors.",
	})
	register(ErrorCodeImportWithIllegalDecodeTemplate, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，导入文件不符合解析模板配置，导入失败。",
		EN:    "Sorry, failed to import data with unsuitable decode template.",
	})
	register(ErrorCodeImportGetPointsFailed, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，导入数据时获取料号检测项信息发生错误。",
		EN:    "Sorry, get detect item failed during import.",
	})
	// 500
	register(ErrorCodeRevertImportFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，无法撤销导入，发生了一些错误。",
		EN:    "Sorry, cannot revert the import with some errors.",
	})
	register(ErrorCodeAssembleDataFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，生成原数据文件时发生错误。",
		EN:    "Sorry, due to some errors, the original data file could not be generated.",
	})
}
