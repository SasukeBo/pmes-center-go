package router

import (
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
)

func authenticate() gin.HandlerFunc {
	return func(gc *gin.Context) {
		token, err := gc.Cookie("access_token")
		if err != nil {
			errormap.SendHttpError(gc, errormap.ErrorCodeUnauthenticated, err)
			return
		}
		user := &orm.User{}
		if err := user.GetWithToken(token); err != nil {
			errormap.SendHttpError(gc, errormap.ErrorCodeUnauthenticated, err)
			return
		}

		gc.Set("current_user", user)
		return
	}
}
