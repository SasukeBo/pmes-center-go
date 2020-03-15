package logic

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/SasukeBo/ftpviewer/orm"
)

func getGinContext(ctx context.Context) *gin.Context {
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
	gc := getGinContext(ctx)

	token, err := gc.Cookie("access_token")
	if err != nil {
		return &gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "get access_token failed.",
			},
		}
	}

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
	gc := getGinContext(ctx)
	user, ok := gc.Get("current_user")
	if !ok {
		return nil
	}

	if u, ok := user.(orm.User); ok {
		return &u
	}

	return nil
}
