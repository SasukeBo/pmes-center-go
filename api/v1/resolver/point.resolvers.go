package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/pmes-data-center/api/v1/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
)

func (r *pointResolver) Material(ctx context.Context, obj *model.Point) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID), nil
}

// Point returns generated.PointResolver implementation.
func (r *Resolver) Point() generated.PointResolver { return &pointResolver{r} }

type pointResolver struct{ *Resolver }
