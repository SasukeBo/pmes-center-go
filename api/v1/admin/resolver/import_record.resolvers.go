package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *importRecordResolver) Material(ctx context.Context, obj *model.ImportRecord) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID)
}

func (r *importRecordResolver) Device(ctx context.Context, obj *model.ImportRecord) (*model.Device, error) {
	return logic.LoadDevice(ctx, obj.DeviceID)
}

func (r *importRecordResolver) User(ctx context.Context, obj *model.ImportRecord) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID)
}

func (r *importRecordResolver) DecodeTemplate(ctx context.Context, obj *model.ImportRecord) (*model.DecodeTemplate, error) {
	return logic.LoadDecodeTemplate(ctx, obj.DecodeTemplateID)
}

// ImportRecord returns generated.ImportRecordResolver implementation.
func (r *Resolver) ImportRecord() generated.ImportRecordResolver { return &importRecordResolver{r} }

type importRecordResolver struct{ *Resolver }
