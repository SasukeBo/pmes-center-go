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

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, analyzeInput model.AnalyzeMaterialInput) (*model.EchartsResult, error) {
	return logic.AnalyzeMaterial(ctx, analyzeInput)
}

func (r *queryResolver) MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	return logic.MaterialYieldTop(ctx, duration, limit)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *queryResolver) MaterialYieldTopRate(ctx context.Context, limit int) (*model.EchartsResult, error) {
	panic(fmt.Errorf("not implemented"))
}
