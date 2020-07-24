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

func (r *mutationResolver) CreateMaterialVersion(ctx context.Context, input model.MaterialVersionInput) (model.ResponseStatus, error) {
	return logic.CreateMaterialVersion(ctx, input)
}

func (r *mutationResolver) DeleteMaterialVersion(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.DeleteMaterialVersion(ctx, id)
}

func (r *mutationResolver) UpdateMaterialVersion(ctx context.Context, id int, input model.MaterialVersionUpdateInput) (model.ResponseStatus, error) {
	return logic.UpdateMaterialVersion(ctx, id, input)
}

func (r *mutationResolver) ChangeMaterialVersionActive(ctx context.Context, id int, active bool) (model.ResponseStatus, error) {
	return logic.ChangeMaterialVersionActive(ctx, id, active)
}

func (r *mutationResolver) SaveDecodeTemplate(ctx context.Context, input model.DecodeTemplateInput) (model.ResponseStatus, error) {
	return logic.SaveDecodeTemplate(ctx, input)
}

func (r *mutationResolver) DeleteDecodeTemplate(ctx context.Context, id int) (model.ResponseStatus, error) {
	return logic.DeleteDecodeTemplate(ctx, id)
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

func (r *mutationResolver) ImportData(ctx context.Context, materialID int, deviceID int, fileTokens []string) (model.ResponseStatus, error) {
	return logic.ImportData(ctx, materialID, deviceID, fileTokens)
}

func (r *mutationResolver) AddUser(ctx context.Context, input model.AddUserInput) (model.ResponseStatus, error) {
	return logic.AddUser(ctx, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) ChangeDefaultTemplate(ctx context.Context, id int, isDefault bool) (model.ResponseStatus, error) {
	return logic.ChangeDefaultTemplate(ctx, id, isDefault)
}
