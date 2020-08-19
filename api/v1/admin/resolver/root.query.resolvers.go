package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	return logic.CurrentUser(ctx)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return logic.Users(ctx)
}

func (r *queryResolver) Materials(ctx context.Context, pattern *string, page int, limit int) (*model.MaterialWrap, error) {
	return logic.Materials(ctx, pattern, page, limit)
}

func (r *queryResolver) Material(ctx context.Context, id int) (*model.Material, error) {
	return logic.Material(ctx, id)
}

func (r *queryResolver) MaterialVersions(ctx context.Context, id int) ([]*model.MaterialVersion, error) {
	return logic.MaterialVersions(ctx, id)
}

func (r *queryResolver) ListMaterialPoints(ctx context.Context, materialVersionID int) ([]*model.Point, error) {
	return logic.ListMaterialPoints(ctx, materialVersionID)
}

func (r *queryResolver) ImportRecords(ctx context.Context, materialVersionID int, deviceID *int, page int, limit int, search model.ImportRecordSearch) (*model.ImportRecordsWrap, error) {
	return logic.ImportRecords(ctx, materialVersionID, deviceID, page, limit, search)
}

func (r *queryResolver) MyImportRecords(ctx context.Context, page int, limit int) (*model.ImportRecordsWrap, error) {
	return logic.MyImportRecords(ctx, page, limit)
}

func (r *queryResolver) ImportStatus(ctx context.Context, id int) (*model.ImportStatusResponse, error) {
	return logic.ImportStatus(ctx, id)
}

func (r *queryResolver) DownloadImportRecords(ctx context.Context, ids []int) (string, error) {
	return logic.DownloadImportRecords(ctx, ids)
}

func (r *queryResolver) ListDecodeTemplate(ctx context.Context, materialID int) ([]*model.DecodeTemplate, error) {
	return logic.ListDecodeTemplate(ctx, materialID)
}

func (r *queryResolver) DecodeTemplateWithVersionID(ctx context.Context, id int) (*model.DecodeTemplate, error) {
	return logic.DecodeTemplateWithVersionID(ctx, id)
}

func (r *queryResolver) ListBarCodeRules(ctx context.Context, search *string, limit int, page int) (*model.BarCodeRuleWrap, error) {
	return logic.ListBarCodeRules(ctx, search, limit, page)
}

func (r *queryResolver) GetBarCodeRule(ctx context.Context, id int) (*model.BarCodeRule, error) {
	return logic.GetBarCodeRule(ctx, id)
}

func (r *queryResolver) ListDevices(ctx context.Context, pattern *string, materialID *int, page int, limit int) (*model.DeviceWrap, error) {
	return logic.ListDevices(ctx, pattern, materialID, page, limit)
}

func (r *queryResolver) Device(ctx context.Context, id int) (*model.Device, error) {
	return logic.Device(ctx, id)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
