package errormap

import "net/http"

// ErrorCode E0006

const (
	// 500
	ErrorCodeRevertImportFailed = "E0006S0500N0001"
	// 400
	ErrorCodeImportGetPointsFailed           = "E0006S0400N0001"
	ErrorCodeImportWithIllegalDecodeTemplate = "E0006S0400N0002"
	ErrorCodeImportFailedWithPanic           = "E0006S0400N0003"
)

func init() {
	// 400
	register(ErrorCodeImportFailedWithPanic, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，导入数据时发生了未知错误。",
		EN:    "Sorry, failed to import data with unknown some .",
	})
	register(ErrorCodeImportWithIllegalDecodeTemplate, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，导入数据时使用的解析模板不合法，导入失败。",
		EN:    "Sorry, failed to import data with illegal decode template.",
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
}
