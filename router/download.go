package router

import (
	"github.com/SasukeBo/configer"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

const xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

func download(c *gin.Context) {
	fileName, ok := c.GetQuery("file_name")
	if !ok {
		c.JSON(http.StatusBadRequest, object{
			"message": "请求参数缺少文件名",
		})
		return
	}

	filePath := filepath.Join(configer.GetString("file_cache_path"), fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, object{
			"message": "下载文件失败",
			"err":     err.Error(),
		})
	}
	os.Remove(filePath) // 删除临时文件
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, xlsxContentType, data)
}
