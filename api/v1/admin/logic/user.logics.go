package logic

import (
	"context"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
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

func CurrentUser(ctx context.Context) (*model.User, error) {
	user := api.CurrentUser(ctx)
	if user == nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeUnauthenticated, errormap.NewOrigin("user not found in cache"))
	}

	var out model.User
	if err := copier.Copy(&out, user); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "user")
	}

	return &out, nil
}

func Users(ctx context.Context) ([]*model.User, error) {
	user := api.CurrentUser(ctx)
	if user == nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeUnauthenticated, errormap.NewOrigin("user not found in cache"))
	}

	var users []orm.User
	if err := orm.Model(&orm.User{}).Find(&users).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "users")
	}

	var outs []*model.User
	for _, u := range users {
		var out model.User
		if err := copier.Copy(&out, &u); err != nil {
			log.Errorln(err)
			continue
		}

		outs = append(outs, &out)
	}

	return outs, nil
}
