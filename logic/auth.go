package logic

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// GetGinContext _
func GetGinContext(ctx context.Context) *gin.Context {
	c := ctx.Value("GinContext")
	if c == nil {
		panic(&gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "get GinContext from context.Context failed.",
			},
		})
	}

	gc, ok := c.(*gin.Context)
	if !ok {
		panic(&gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "assert GinContext failed.",
			},
		})
	}

	return gc
}

// Authenticate _
func Authenticate(ctx context.Context) error {
	gc := GetGinContext(ctx)

	token := gc.GetHeader("Access-Token")
	user := orm.GetUserWithTokenCache(token)
	if user == nil {
		return &gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "get user by access_token failed.",
			},
		}
	}

	gc.Set("current_user", *user)

	return nil
}

// CurrentUser _
func CurrentUser(ctx context.Context) *orm.User {
	gc := GetGinContext(ctx)
	user, ok := gc.Get("current_user")
	if !ok {
		return nil
	}

	if u, ok := user.(orm.User); ok {
		return &u
	}

	return nil
}

// GenToken _
func GenToken(base string) string {
	t := time.Now()
	data := []byte(base + t.String())
	return base64.StdEncoding.EncodeToString(data)
}
