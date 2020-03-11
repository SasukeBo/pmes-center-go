package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/SasukeBo/ftpviewer/graph/model"
)

func (r *mutationResolver) Login(ctx context.Context, loginInput model.LoginInput) (*model.User, error) {
	user := model.User{}
	user.ID = 1
	user.Account = loginInput.Account
	user.Password = loginInput.Password
	return &user, nil
}

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
