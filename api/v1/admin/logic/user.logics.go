package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
)

func LoadUser(ctx context.Context, userID uint) *model.User {
	var user orm.User
	if err := user.Get(userID); err != nil {
		return nil
	}
	var out model.User
	if err := copier.Copy(&out, &user); err != nil {
		return nil
	}

	return &out
}
