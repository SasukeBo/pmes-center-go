package errormap

import "net/http"

// ErrorCode E0008

const (
	// 400
	ErrorCodeBarCodeReservedKey = "E0008S0400N0001"
)

func init() {
	// 400
	register(ErrorCodeBarCodeReservedKey, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，{{.field_1}}为保留字段，不能作为解析项的Key值，请更换。",
		EN:    "Sorry, {{.field_1}} is a reserved key and cannot be used by decoding items. Please change another one.",
	})
}
