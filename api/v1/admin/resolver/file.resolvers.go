package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *fileResolver) User(ctx context.Context, obj *model.File) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID), nil
}

// File returns generated.FileResolver implementation.
func (r *Resolver) File() generated.FileResolver { return &fileResolver{r} }

type fileResolver struct{ *Resolver }
