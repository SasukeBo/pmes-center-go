package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *materialVersionResolver) Material(ctx context.Context, obj *model.MaterialVersion) (*model.Material, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *materialVersionResolver) User(ctx context.Context, obj *model.MaterialVersion) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// MaterialVersion returns generated.MaterialVersionResolver implementation.
func (r *Resolver) MaterialVersion() generated.MaterialVersionResolver {
	return &materialVersionResolver{r}
}

type materialVersionResolver struct{ *Resolver }
