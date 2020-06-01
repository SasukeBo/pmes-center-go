package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/jinzhu/copier"
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
