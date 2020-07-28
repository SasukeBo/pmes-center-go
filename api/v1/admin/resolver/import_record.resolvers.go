package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
)

func (r *importRecordResolver) File(ctx context.Context, obj *model.ImportRecord) (*model.File, error) {
	if obj.FileID == nil {
		return nil, nil
	}

	return logic.LoadFile(ctx, *obj.FileID), nil
}

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

func (r *importRecordResolver) MaterialVersion(ctx context.Context, obj *model.ImportRecord) (*model.MaterialVersion, error) {
	return logic.LoadMaterialVersion(ctx, obj.MaterialVersionID), nil
}

// ImportRecord returns generated.ImportRecordResolver implementation.
func (r *Resolver) ImportRecord() generated.ImportRecordResolver { return &importRecordResolver{r} }

type importRecordResolver struct{ *Resolver }
