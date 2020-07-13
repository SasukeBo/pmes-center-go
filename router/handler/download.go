package handler

import (
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type object map[string]string

func DownloadXlsxFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileToken, ok := c.GetQuery("file_token")
		if !ok {
			errormap.SendHttpError(
				c, errormap.ErrorCodeRequestInputMissingFieldError,
				errormap.NewOrigin("request missing file_token in query"),
				"file_token",
			)
			return
		}

		var file orm.File
		if err := file.GetByToken(fileToken); err != nil {
			errormap.SendHttpError(c, err.GetCode(), err, "file")
			return
		}

		filePath := filepath.Join(configer.GetString("file_cache_path"), file.Path)
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeFileDownloadError, err)
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+file.Name)
		c.Data(http.StatusOK, file.ContentType, data)
	}
}
