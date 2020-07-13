package errormap

import "net/http"

// ErrorCode E0004

const (
	// 400
	ErrorCodeFileHandleError     = "E0004S0400N0001"
	ErrorCodeFileExtensionError  = "E0004S0400N0002"
	ErrorCodeFileUploadError     = "E0004S0400N0003"
	ErrorCodeFileOpenFailedError = "E0004S0400N0004"

	// 500
	ErrorCodeFileDownloadError = "E0004S0500N0001"
)

func init() {
	// 400
	register(ErrorCodeFileHandleError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，无法处理该文件，请确认文件是否已损坏。",
		EN:    "Sorry, cannot handle the file, please check whether the file is damaged.",
	})
	register(ErrorCodeFileExtensionError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，文件格式不正确，需要{{.field_1}}文件。",
		EN:    "Sorry, the file extension is wrong, {{.field_1}} file in need.",
	})
	register(ErrorCodeFileUploadError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，文件上传失败，发生了一些错误。",
		EN:    "Sorry, failed to upload the file with some errors.",
	})
	register(ErrorCodeFileOpenFailedError, http.StatusBadRequest, langMap{
		ZH_CN: "对不起，打开文件时发生了一些错误。",
		EN:    "Sorry, failed to open the file with some errors.",
	})

	// 500
	register(ErrorCodeFileDownloadError, http.StatusInternalServerError, langMap{
		ZH_CN: "对不起，无法下载该文件，发生了一些错误。",
		EN:    "Sorry, failed to download the file with some errors.",
	})
}
