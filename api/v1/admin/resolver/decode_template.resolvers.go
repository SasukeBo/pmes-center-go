package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *decodeTemplateResolver) Material(ctx context.Context, obj *model.DecodeTemplate) (*model.Material, error) {
	return logic.LoadMaterial(ctx, obj.MaterialID), nil
}

func (r *decodeTemplateResolver) MaterialVersion(ctx context.Context, obj *model.DecodeTemplate) (*model.MaterialVersion, error) {
	return logic.LoadMaterialVersion(ctx, obj.MaterialVersionID), nil
}

func (r *decodeTemplateResolver) User(ctx context.Context, obj *model.DecodeTemplate) (*model.User, error) {
	return logic.LoadUser(ctx, obj.UserID), nil
}

func (r *decodeTemplateResolver) BarCodeRule(ctx context.Context, obj *model.DecodeTemplate) (*model.BarCodeRule, error) {
	return logic.LoadBarCodeRule(ctx, obj.BarCodeRuleID), nil
}

// DecodeTemplate returns generated.DecodeTemplateResolver implementation.
func (r *Resolver) DecodeTemplate() generated.DecodeTemplateResolver {
	return &decodeTemplateResolver{r}
}

type decodeTemplateResolver struct{ *Resolver }
