package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/SasukeBo/pmes-data-center/api/v1/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
)

func (r *queryResolver) Materials(ctx context.Context, search *string, page int, limit int) (*model.MaterialsWrap, error) {
	return logic.Materials(ctx, search, page, limit)
}

func (r *queryResolver) Material(ctx context.Context, id int) (*model.Material, error) {
	return logic.Material(ctx, id)
}

func (r *queryResolver) MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	return logic.MaterialYieldTop(ctx, duration, limit)
}

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, id int, deviceID *int, versionID *int, duration []*time.Time) (*model.Material, error) {
	return logic.AnalyzeMaterial(ctx, id, deviceID, versionID, duration)
}

func (r *queryResolver) GroupAnalyzeMaterial(ctx context.Context, analyzeInput model.GraphInput, versionID *int) (*model.EchartsResult, error) {
	return logic.GroupAnalyzeMaterial(ctx, analyzeInput, versionID)
}

func (r *queryResolver) ProductAttributes(ctx context.Context, materialID int, versionID *int) ([]*model.ProductAttribute, error) {
	return logic.ProductAttributes(ctx, materialID, versionID)
}

func (r *queryResolver) MaterialVersions(ctx context.Context, id int, search *string, limit *int, isActive *bool) ([]*model.MaterialVersion, error) {
	return logic.MaterialVersions(ctx, id, search, limit, isActive)
}

func (r *queryResolver) MaterialVersion(ctx context.Context, id int) (*model.MaterialVersion, error) {
	return logic.MaterialVersion(ctx, id)
}

func (r *queryResolver) MaterialActiveVersion(ctx context.Context, id int) (*model.MaterialVersion, error) {
	return logic.MaterialActiveVersion(ctx, id)
}

func (r *queryResolver) Device(ctx context.Context, id int) (*model.Device, error) {
	return logic.Device(ctx, id)
}

func (r *queryResolver) Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	return logic.Devices(ctx, materialID)
}

func (r *queryResolver) AnalyzeDevices(ctx context.Context, materialID int, versionID *int) ([]*model.DeviceResult, error) {
	return logic.AnalyzeDevices(ctx, materialID, versionID)
}

func (r *queryResolver) AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	return logic.AnalyzeDevice(ctx, searchInput)
}

func (r *queryResolver) GroupAnalyzeDevice(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return logic.GroupAnalyzeDevice(ctx, analyzeInput)
}

func (r *queryResolver) Point(ctx context.Context, id int) (*model.Point, error) {
	return logic.Point(ctx, id)
}

func (r *queryResolver) SizeUnYieldTop(ctx context.Context, groupInput model.GraphInput, versionID *int) (*model.EchartsResult, error) {
	return logic.SizeUnYieldTop(ctx, groupInput, versionID)
}

func (r *queryResolver) PointListWithYield(ctx context.Context, materialID int, versionID *int, limit int, page int) (*model.PointListWithYieldResponse, error) {
	return logic.PointListWithYield(ctx, materialID, versionID, limit, page)
}

func (r *queryResolver) SizeNormalDistribution(ctx context.Context, id int, duration []*time.Time, filters map[string]interface{}) (*model.PointResult, error) {
	return logic.SizeNormalDistribution(ctx, id, duration, filters)
}

func (r *queryResolver) GroupAnalyzePoint(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return logic.GroupAnalyzePoint(ctx, analyzeInput)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
