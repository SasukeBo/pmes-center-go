package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *decodeTemplateResolver) Material(ctx context.Context, obj *model.DecodeTemplate) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID)
}

func (r *decodeTemplateResolver) User(ctx context.Context, obj *model.DecodeTemplate) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID)
}

// DecodeTemplate returns generated.DecodeTemplateResolver implementation.
func (r *Resolver) DecodeTemplate() generated.DecodeTemplateResolver {
	return &decodeTemplateResolver{r}
}

type decodeTemplateResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *decodeTemplateResolver) CreatedAtColumnIndex(ctx context.Context, obj *model.DecodeTemplate) (string, error) {
	panic(fmt.Errorf("not implemented"))
}
