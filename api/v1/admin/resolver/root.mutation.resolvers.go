package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
)

func (r *mutationResolver) AddMaterial(ctx context.Context, input model.MaterialCreateInput) (*model.Material, error) {
	return logic.AddMaterial(ctx, input)
}

func (r *mutationResolver) DeleteMaterial(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.DeleteMaterial(ctx, id)
}

func (r *mutationResolver) UpdateMaterial(ctx context.Context, input model.MaterialUpdateInput) (*model.Material, error) {
	return logic.UpdateMaterial(ctx, input)
}

func (r *mutationResolver) MaterialFetch(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.MaterialFetch(ctx, id)
}

func (r *mutationResolver) SaveMaterialVersion(ctx context.Context, input model.MaterialVersionInput) (model.ResponseStatus, error) {
	return logic.SaveMaterialVersion(ctx, input)
}

func (r *mutationResolver) SaveDecodeTemplate(ctx context.Context, input model.DecodeTemplateInput) (*model.DecodeTemplate, error) {
	return logic.SaveDecodeTemplate(ctx, input)
}

func (r *mutationResolver) DeleteDecodeTemplate(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.DeleteDecodeTemplate(ctx, id)
}

func (r *mutationResolver) ChangeDefaultTemplate(ctx context.Context, id int, isDefault bool) (model.ResponseStatus, error) {
	return logic.ChangeDefaultTemplate(ctx, id, isDefault)
}

func (r *mutationResolver) ParseImportPoints(ctx context.Context, file graphql.Upload) ([]*model.Point, error) {
	return logic.ParseImportPoints(ctx, file)
}

func (r *mutationResolver) SavePoints(ctx context.Context, materialID int, saveItems []*model.PointCreateInput, deleteItems []int) (model.ResponseStatus, error) {
	return logic.SavePoints(ctx, materialID, saveItems, deleteItems)
}

func (r *mutationResolver) SaveDevice(ctx context.Context, input model.DeviceInput) (*model.Device, error) {
	return logic.SaveDevice(ctx, input)
}

func (r *mutationResolver) DeleteDevice(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.DeleteDevice(ctx, id)
}

func (r *mutationResolver) RevertImports(ctx context.Context, ids []int) (model.ResponseStatus, error) {
	return logic.RevertImports(ctx, ids)
}

func (r *mutationResolver) ToggleBlockImports(ctx context.Context, ids []int, block bool) (model.ResponseStatus, error) {
	return logic.ToggleBlockImports(ctx, ids, block)
}

func (r *mutationResolver) ImportData(ctx context.Context, materialID int, deviceID int, decodeTemplateID int, fileTokens []string) (model.ResponseStatus, error) {
	return logic.ImportData(ctx, materialID, deviceID, decodeTemplateID, fileTokens)
}

func (r *mutationResolver) AddUser(ctx context.Context, input model.AddUserInput) (model.ResponseStatus, error) {
	return logic.AddUser(ctx, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
