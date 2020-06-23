package handler

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func Post() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("hello world")
		iCurrentUser, ok := c.Get("current_user")
		if !ok {
			errormap.SendHttpError(c, errormap.ErrorCodeUnauthenticated, nil)
			return
		}
		currentUser := iCurrentUser.(orm.User)

		post, err := c.FormFile("file")
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeFileUploadError, err)
			return
		}

		dst := configer.GetString("file_cache_path")
		path := filepath.Join(dst, "post")
		err = c.SaveUploadedFile(post, path)
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeFileUploadError, err)
			return
		}

		fmt.Println(post.Filename, post.Size, post.Header)

		file := orm.File{
			Name:        post.Filename,
			Path:        filepath.Join(path, post.Filename),
			UserID:      currentUser.ID,
			Size:        uint(post.Size),
			ContentType: post.Header["Content-Type"][0],
		}
		err = orm.Create(&file).Error
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeCreateObjectError, err, "file")
			return
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"token": file.Token,
		})
		return
	}
}
