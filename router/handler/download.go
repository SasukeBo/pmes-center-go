package handler

import (
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
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

		var filePath string
		if strings.HasPrefix(file.Path, "/") {
			filePath = file.Path
		} else {
			dst := configer.GetString("file_cache_path")
			filePath = filepath.Join(dst, file.Path)
		}

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeFileDownloadError, err)
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+file.Name)
		c.Data(http.StatusOK, file.ContentType, data)
	}
}
