package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *materialVersionResolver) Material(ctx context.Context, obj *model.MaterialVersion) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID), nil
}

func (r *materialVersionResolver) User(ctx context.Context, obj *model.MaterialVersion) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID), nil
}

// MaterialVersion returns generated.MaterialVersionResolver implementation.
func (r *Resolver) MaterialVersion() generated.MaterialVersionResolver {
	return &materialVersionResolver{r}
}

type materialVersionResolver struct{ *Resolver }
