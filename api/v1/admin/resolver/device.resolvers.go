package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *deviceResolver) Material(ctx context.Context, obj *model.Device) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID), nil
}

// Device returns generated.DeviceResolver implementation.
func (r *Resolver) Device() generated.DeviceResolver { return &deviceResolver{r} }

type deviceResolver struct{ *Resolver }
