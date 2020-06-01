package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

func CurrentUser(ctx context.Context) (*model.User, error) {
	gc := getGinContext(ctx)
	user := currentUser(gc)
	if user == nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeUnauthenticated, nil)
	}

	var out model.User
	if err := copier.Copy(&out, user); err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "user")
	}

	return &out, nil
}

func Login(ctx context.Context, loginInput model.LoginInput) (*model.User, error) {
	gc := getGinContext(ctx)
	var user orm.User

	if err := orm.DB.Where("username = ? AND password = ?", loginInput.Account, util.Encrypt(loginInput.Password)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeAccountPasswordIncorrect, err)
		}

		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeLoginFailed, err)
	}

	token := genToken(user.Password)
	if err := orm.DB.Model(&user).Update("access_token", token).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeLoginFailed, err)
	}

	maxAge := 7 * 24 * 60 * 60
	gc.SetCookie("access_token", token, maxAge, "/", "", false, true)

	var out model.User
	if err := copier.Copy(&out, &user); err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "user")
	}

	return &out, nil
}

func Logout(ctx context.Context) (string, error) {
	gc := getGinContext(ctx)
	user := currentUser(gc)
	if user == nil {
		return "error", errormap.SendGQLError(gc, errormap.ErrorCodeUnauthenticated, errormap.NewOrigin("current user is nil"))
	}

	user.AccessToken = ""
	if err := orm.DB.Save(user).Error; err != nil {
		return "error", errormap.SendGQLError(gc, errormap.ErrorCodeLogoutFailed, errormap.NewOrigin("clean user access token failed with error: %v", err))
	}

	return "ok", nil
}
