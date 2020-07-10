package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"time"

	"github.com/SasukeBo/ftpviewer/api/v1/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
)

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	return logic.CurrentUser(ctx)
}

func (r *queryResolver) Materials(ctx context.Context, search *string, page int, limit int) (*model.MaterialsWrap, error) {
	return logic.Materials(ctx, search, page, limit)
}

func (r *queryResolver) Material(ctx context.Context, id int) (*model.Material, error) {
	return logic.Material(ctx, id)
}

func (r *queryResolver) MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	return logic.MaterialYieldTop(ctx, duration, limit)
}

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, searchInput model.Search) (*model.MaterialResult, error) {
	return logic.AnalyzeMaterial(ctx, searchInput)
}

func (r *queryResolver) GroupAnalyzeMaterial(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return logic.GroupAnalyzeMaterial(ctx, analyzeInput)
}

func (r *queryResolver) ProductAttributes(ctx context.Context, materialID int) ([]*model.ProductAttribute, error) {
	return logic.ProductAttributes(ctx, materialID)
}

func (r *queryResolver) Device(ctx context.Context, id int) (*model.Device, error) {
	return logic.Device(ctx, id)
}

func (r *queryResolver) Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	return logic.Devices(ctx, materialID)
}

func (r *queryResolver) AnalyzeDevices(ctx context.Context, materialID int) ([]*model.DeviceResult, error) {
	return logic.AnalyzeDevices(ctx, materialID)
}

func (r *queryResolver) AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	return logic.AnalyzeDevice(ctx, searchInput)
}

func (r *queryResolver) GroupAnalyzeDevice(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return logic.GroupAnalyzeDevice(ctx, analyzeInput)
}

func (r *queryResolver) SizeUnYieldTop(ctx context.Context, groupInput model.GraphInput) (*model.EchartsResult, error) {
	return logic.SizeUnYieldTop(ctx, groupInput)
}

func (r *queryResolver) PointListWithYield(ctx context.Context, materialID int, limit int, page int) (*model.PointListWithYieldResponse, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
