package logic

import (
	"context"
	"encoding/base64"
	"github.com/SasukeBo/ftpviewer/errormap"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
)

// GetGinContext _
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

//func ValidateExpired() error {
//	expiredConfig := orm.GetSystemConfig("expired_at")
//	if expiredConfig == nil {
//		return &gqlerror.Error{
//			Message: "System not active.",
//			Extensions: map[string]interface{}{
//				"originErr": "expired config not found.",
//			},
//		}
//	}
//
//	if expiredConfig.Value != "unlimited" {
//		expiredAt, err := time.Parse(time.RFC3339, expiredConfig.Value)
//		if err != nil {
//			return &gqlerror.Error{
//				Message: "System not active.",
//				Extensions: map[string]interface{}{
//					"originErr": err.Error(),
//				},
//			}
//		}
//
//		if time.Now().After(expiredAt) {
//			return &gqlerror.Error{
//				Message: "System not active.",
//				Extensions: map[string]interface{}{
//					"originErr": "system expired at " + expiredConfig.Value,
//				},
//			}
//		}
//	}
//
//	return nil
//}

// Authenticate _
func Authenticate(ctx context.Context) error {
	gc := GetGinContext(ctx)
	token, err := gc.Cookie("access_token")
	if err != nil {
		return err
	}
	user := orm.GetUserWithToken(token)
	if user == nil {
		return errormap.NewOrigin("get user by access_token failed.")
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
