package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/jinzhu/copier"
)

func CurrentUser(ctx context.Context) (*model.User, error) {
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if user == nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeUnauthenticated, nil)
	}

	var out model.User
	if err := copier.Copy(&out, user); err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "user")
	}

	return &out, nil
}
