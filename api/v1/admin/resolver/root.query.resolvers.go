package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *queryResolver) Materials(ctx context.Context, pattern *string, page int, limit int) (*model.MaterialWrap, error) {
	return logic.Materials(ctx, pattern, page, limit)
}

func (r *queryResolver) ListMaterialPoints(ctx context.Context, materialID int, page int, limit int) (*model.PointWrap, error) {
	return logic.ListMaterialPoints(ctx, materialID, page, limit)
}

func (r *queryResolver) ImportRecords(ctx context.Context, materialID int, deviceID *int, page int, limit int) (*model.ImportRecordsWrap, error) {
	return logic.ImportRecords(ctx, materialID, deviceID, page, limit)
}

func (r *queryResolver) ListDecodeTemplate(ctx context.Context, materialID int) ([]*model.DecodeTemplate, error) {
	return logic.ListDecodeTemplate(ctx, materialID)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
