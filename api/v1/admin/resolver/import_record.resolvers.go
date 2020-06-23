package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
)

func (r *importRecordResolver) Material(ctx context.Context, obj *model.ImportRecord) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID), nil
}

func (r *importRecordResolver) Device(ctx context.Context, obj *model.ImportRecord) (*model.Device, error) {
	return logic.LoadDevice(ctx, obj.DeviceID), nil
}

func (r *importRecordResolver) ErrorMessage(ctx context.Context, obj *model.ImportRecord) (*string, error) {
	if obj.ErrorCode != nil {
		message := errormap.DecodeError(ctx, *obj.ErrorCode)
		return &message, nil
	}

	return nil, nil
}

func (r *importRecordResolver) User(ctx context.Context, obj *model.ImportRecord) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID), nil
}

func (r *importRecordResolver) DecodeTemplate(ctx context.Context, obj *model.ImportRecord) (*model.DecodeTemplate, error) {
	return logic.LoadDecodeTemplate(ctx, obj.DecodeTemplateID), nil
}

// ImportRecord returns generated.ImportRecordResolver implementation.
func (r *Resolver) ImportRecord() generated.ImportRecordResolver { return &importRecordResolver{r} }

type importRecordResolver struct{ *Resolver }
