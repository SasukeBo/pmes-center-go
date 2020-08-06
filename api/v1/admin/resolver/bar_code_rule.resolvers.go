package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *barCodeRuleResolver) User(ctx context.Context, obj *model.BarCodeRule) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID), nil
}

// BarCodeRule returns generated.BarCodeRuleResolver implementation.
func (r *Resolver) BarCodeRule() generated.BarCodeRuleResolver { return &barCodeRuleResolver{r} }

type barCodeRuleResolver struct{ *Resolver }
