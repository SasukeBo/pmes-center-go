package logic

import (
	"context"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/util"
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

func AddUser(ctx context.Context, input model.AddUserInput) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if user == nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeUnauthenticated, errormap.NewOrigin("user not found in cache"))
	}

	var newUser = orm.User{
		Name:     input.Name,
		IsAdmin:  input.IsAdmin,
		Account:  input.Account,
		Password: util.Encrypt(input.Password),
	}

	if err := orm.Create(&newUser).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "user")
	}

	return model.ResponseStatusOk, nil
}
