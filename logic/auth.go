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
func GetGinContext(ctx context.Context) (*gin.Context, error) {
	c := ctx.Value("GinContext")
	if c == nil {
		return nil, &gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "get GinContext from context.Context failed.",
			},
		}
	}

	gc, ok := c.(*gin.Context)
	if !ok {
		return nil, &gqlerror.Error{
			Message: "用户验证失败",
			Extensions: map[string]interface{}{
				"originErr": "assert GinContext failed.",
			},
		}
	}

	return gc, nil
}

func ValidateExpired() error {
	expiredConfig := orm.GetSystemConfig("expired_at")
	if expiredConfig == nil {
		return &gqlerror.Error{
			Message: "System not active.",
			Extensions: map[string]interface{}{
				"originErr": "expired config not found.",
			},
		}
	}

	if expiredConfig.Value != "unlimited" {
		expiredAt, err := time.Parse(time.RFC3339, expiredConfig.Value)
		if err != nil {
			return &gqlerror.Error{
				Message: "System not active.",
				Extensions: map[string]interface{}{
					"originErr": err.Error(),
				},
			}
		}

		if time.Now().After(expiredAt) {
			return &gqlerror.Error{
				Message: "System not active.",
				Extensions: map[string]interface{}{
					"originErr": "system expired at " + expiredConfig.Value,
				},
			}
		}
	}

	return nil
}

// Authenticate _
func Authenticate(ctx context.Context) error {
	return nil
	gc, err := GetGinContext(ctx)
	if err != nil {
		return err
	}

	token := gc.GetHeader("Access-Token")
	user := orm.GetUserWithToken(token)
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
	gc, err := GetGinContext(ctx)
	if err != nil {
		return nil
	}

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
