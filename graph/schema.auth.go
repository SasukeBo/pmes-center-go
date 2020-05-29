package graph

import (
	"context"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/graph/logic"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/jinzhu/gorm"
)

func (r *mutationResolver) Login(ctx context.Context, loginInput model.LoginInput) (*model.User, error) {
	gc := logic.GetGinContext(ctx)
	var user orm.User

	if err := orm.DB.Where("username = ? AND password = ?", loginInput.Account, util.Encrypt(loginInput.Password)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeAccountPasswordIncorrect, err)
		}

		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeLoginFailed, err)
	}

	token := logic.GenToken(user.Password)
	if err := orm.DB.Model(&user).Update("access_token", token).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeLoginFailed, err)
	}

	maxAge := 7 * 24 * 60 * 60
	gc.SetCookie("access_token", token, maxAge, "/", "", false, true)

	userID := int(user.ID)
	return &model.User{
		ID:      &userID,
		Account: &user.Username,
		Admin:   &user.Admin,
	}, nil
}

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	gc := logic.GetGinContext(ctx)
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	user := logic.CurrentUser(ctx)
	if user == nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeUnauthenticated, errormap.NewOrigin("current user is nil"))
	}

	return &model.User{
		ID:      intP(int(user.ID)),
		Account: &user.Username,
		Admin:   &user.Admin,
	}, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (string, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return "error", err
	}

	user := logic.CurrentUser(ctx)
	if user == nil {
		return "error", NewGQLError("用户未登录", "current user is nil")
	}

	user.AccessToken = ""
	if err := orm.DB.Save(user).Error; err != nil {
		return "error", NewGQLError("退出登录失败，发生了一些错误", err.Error())
	}

	return "ok", nil
}
