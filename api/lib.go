package api

import (
	"context"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/log"
	"github.com/gin-gonic/gin"
)

func getGinContext(ctx context.Context) *gin.Context {
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

func CurrentUser(ctx context.Context) *orm.User {
	gc := getGinContext(ctx)
	user, ok := gc.Get("current_user")
	if !ok {
		log.Warn("current user not found in gin.Context")
		return nil
	}

	if u, ok := user.(orm.User); ok {
		return &u
	}

	return nil
}
