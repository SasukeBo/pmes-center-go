package errormap

import "net/http"

// ErrorCode E0000

const (
	// 400
	ErrorCodeCreateFailedError             = "E0000S0400N0001"
	ErrorCodeRequestInputMissingFieldError = "E0000S0400N0002"
	ErrorCodeGetObjectFailed               = "E0000S0400N0003"
	ErrorCodeBadRequestParams              = "E0000S0400N0004"
	// 404
	ErrorCodeObjectNotFound = "E0000S0404N0001"
	// 500
	ErrorCodeInternalError       = "E0000S0500N0001"
	ErrorCodeSaveObjectError     = "E0000S0500N0002"
	ErrorCodeDeleteObjectError   = "E0000S0500N0003"
	ErrorCodeCreateObjectError   = "E0000S0500N0004"
	ErrorCodeTransferObjectError = "E0000S0500N0005"
	ErrorCodeCountObjectFailed   = "E0000S0500N0006"
)

func init() {
	// 400
	register(ErrorCodeBadRequestParams, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，查询参数不合法，请检查您的输入。",
		EN:    "Sorry, the search params are illegal, please check your input.",
	})
	register(ErrorCodeGetObjectFailed, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，查找{{.field_1}}失败，发生了一些错误。",
		EN:    "Sorry, failed to get {{.field_1}} with some errors.",
	})
	register(ErrorCodeRequestInputMissingFieldError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，请求参数错误，缺少{{.field_1}}。",
		EN:    "Sorry, the request input variables missing {{.field_1}}.",
	})
	register(ErrorCodeCreateFailedError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，创建{{.field_1}}时发生错误，请重试。",
		EN:    "Sorry, create {{.field_1}} failed with some error, please try again.",
	})

	// 404
	register(ErrorCodeObjectNotFound, http.StatusNotFound, langMap{
		ZH_CN: "对不起，没有找到该{{.field_1}}，请确认您的输入。",
		EN:    "Sorry, cannot find the {{.field_1}}, please check your input.",
	})

	// 500
	register(ErrorCodeDeleteObjectError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，删除{{.field_1}}数据失败，发生了一些错误。",
		EN:    "Sorry, failed to delete the data of {{.field_1}} with some errors.",
	})
	register(ErrorCodeTransferObjectError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，转换{{.field_1}}数据失败，发生了一些错误。",
		EN:    "Sorry, failed to transfer the data of {{.field_1}} with some errors.",
	})
	register(ErrorCodeSaveObjectError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，保存{{.field_1}}数据失败，发生了一些错误。",
		EN:    "Sorry, failed to save the data of {{.field_1}} with some errors.",
	})
	register(ErrorCodeInternalError, http.StatusInternalServerError, langMap{
		ZH_CN: "系统错误。",
		EN:    "Internal error.",
	})
	register(ErrorCodeCountObjectFailed, http.StatusInternalServerError, langMap{
		ZH_CN: "抱歉，统计{{.field_1}}数据总数时，发生了一些错误。",
		EN:    "Sorry, count {{.field_1}} data failed with some errors.",
	})
	register(ErrorCodeCreateObjectError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，创建{{.field_1}}数据失败，发生了一些错误。",
		EN:    "Sorry, failed to create the data of {{.field_1}} with some errors.",
	})
}
