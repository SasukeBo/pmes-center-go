package api

import (
	"context"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/gin-gonic/gin"
)

func GetGinContext(ctx context.Context) *gin.Context {
	c := ctx.Value("GinContext")
	if c == nil {
		panic("gin.Context not found in ctx")
	}

	gc, ok := c.(*gin.Context)
	if !ok {
		panic("GinContext is not a gin.Context")
	}

	return gc
}

func CurrentUser(gCtx *gin.Context) *orm.User {
	user, ok := gCtx.Get("current_user")
	if !ok {
		log.Warn("current user not found in gin.Context")
		return nil
	}

	if u, ok := user.(orm.User); ok {
		return &u
	}

	return nil
}
