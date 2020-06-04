package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api/v1/logic"

	"github.com/SasukeBo/ftpviewer/api/v1/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
)

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	return logic.CurrentUser(ctx)
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
