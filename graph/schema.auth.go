package graph

import (
	"context"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/jinzhu/gorm"
)

func (r *mutationResolver) Login(ctx context.Context, loginInput model.LoginInput) (*model.User, error) {
	var user orm.User

	if err := orm.DB.Where("username = ? AND password = ?", loginInput.Account, util.Encrypt(loginInput.Password)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewGQLError("账号或密码不正确", err.Error())
		}

		return nil, NewGQLError("登录失败", err.Error())
	}

	token := logic.GenToken(user.Password)
	if err := orm.DB.Model(&user).Update("access_token", token).Error; err != nil {
		return nil, NewGQLError("登录失败", err.Error())
	}

	gc, err := logic.GetGinContext(ctx)
	if err != nil {
		return nil, err
	}

	if gc != nil {
		maxAge := 7 * 24 * 60 * 60
		gc.SetCookie("access_token", token, maxAge, "/", "", false, true)
	}

	userID := int(user.ID)
	return &model.User{
		ID:      &userID,
		Account: &user.Username,
		Admin:   &user.Admin,
	}, nil
}

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	user := logic.CurrentUser(ctx)
	if user == nil {
		return nil, NewGQLError("用户未登录", "current user is nil")
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
