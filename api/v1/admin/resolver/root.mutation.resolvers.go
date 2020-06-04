package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/generated"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
)

func (r *mutationResolver) UpdateMaterial(ctx context.Context, input model.MaterialUpdateInput) (*model.Material, error) {
	return nil, nil
	//if err := logic.Authenticate(ctx); err != nil {
	//	return nil, err
	//}
	//
	//var material orm.Material
	//if err := orm.DB.Model(&orm.Material{}).Where("id = ?", input.ID).First(&material).Error; err != nil {
	//	if err == gorm.ErrRecordNotFound {
	//		return nil, NewGQLError("料号不存在", err.Error())
	//	}
	//
	//	return nil, NewGQLError("获取料号失败", err.Error())
	//}
	//
	//if input.ProjectRemark != nil {
	//	material.ProjectRemark = *input.ProjectRemark
	//}
	//
	//if input.CustomerCode != nil {
	//	material.CustomerCode = *input.CustomerCode
	//}
	//
	//if err := orm.DB.Save(&material).Error; err != nil {
	//	return nil, NewGQLError("保存料号失败", err.Error())
	//}
	//
	//out := &model.Material{
	//	ID:            material.ID,
	//	Name:          material.Name,
	//	CustomerCode:  &material.CustomerCode,
	//	ProjectRemark: &material.ProjectRemark,
	//}
	//
	//return out, nil
}

func (r *mutationResolver) DeleteMaterial(ctx context.Context, id int) (string, error) {
	return "", nil
	//if err := logic.Authenticate(ctx); err != nil {
	//	return "error", err
	//}
	//
	//tx := orm.DB.Begin()
	//defer tx.Rollback()
	//
	//var sizeIDs []int
	//if err := tx.Model(&orm.Size{}).Where("material_id = ?", id).Pluck("id", &sizeIDs).Error; err != nil {
	//	return "error", NewGQLError("删除料号尺寸数据失败，删除操作被终止", err.Error())
	//}
	//
	//var pointIDs []int
	//if err := tx.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Pluck("id", &pointIDs).Error; err != nil {
	//	return "error", NewGQLError("删除料号尺寸点位数据失败，删除操作被终止", err.Error())
	//}
	//
	//if err := tx.Where("id = ?", id).Delete(orm.Material{}).Error; err != nil {
	//	return "error", NewGQLError("删除料号失败，发生了一些错误", err.Error())
	//}
	//
	//if err := tx.Where("material_id = ?", id).Delete(orm.File{}).Error; err != nil {
	//	return "error", NewGQLError("删除料号失败，发生了一些错误", err.Error())
	//}
	//
	//if err := tx.Where("material_id = ?", id).Delete(orm.Device{}).Error; err != nil {
	//	return "error", NewGQLError("删除料号设备失败，发生了一些错误", err.Error())
	//}
	//
	//tx.Commit()
	//
	//go func() {
	//	orm.DB.Where("material_id = ?", id).Delete(orm.Product{})
	//	orm.DB.Where("point_id in (?)", pointIDs).Delete(orm.PointValue{})
	//	orm.DB.Where("id in (?)", pointIDs).Delete(orm.Point{})
	//	orm.DB.Where("id in (?)", sizeIDs).Delete(orm.Size{})
	//}()
	//
	//return "料号删除成功", nil
}

func (r *mutationResolver) CancelExport(ctx context.Context, opID string) (string, error) {
	return "", nil
	//err := logic.CancelExport(opID)
	//if err != nil {
	//	return "error", NewGQLError("取消导出失败", err.Error())
	//}
	//
	//return "ok", nil
}

func (r *mutationResolver) ImportPoints(ctx context.Context, file graphql.Upload, materialID int) ([]*model.Point, error) {
	return logic.ImportPoints(ctx, file, materialID)
}

func (r *mutationResolver) Setting(ctx context.Context, settingInput model.SettingInput) (*model.SystemConfig, error) {
	return nil, nil
	//if err := logic.Authenticate(ctx); err != nil {
	//	return nil, err
	//}
	//
	//user := logic.CurrentUser(ctx)
	//if user == nil || !user.Admin {
	//	return nil, NewGQLError("添加系统配置失败，您不是Admin", fmt.Sprintf("%+v", *user))
	//}
	//
	//conf := orm.GetSystemConfig(settingInput.Key)
	//if conf == nil {
	//	conf = &orm.SystemConfig{
	//		Key:   settingInput.Key,
	//		Value: settingInput.Value,
	//	}
	//} else {
	//	conf.Value = settingInput.Value
	//}
	//
	//if err := orm.DB.Save(conf).Error; err != nil {
	//	return nil, NewGQLError("添加系统配置失败", err.Error())
	//}
	//
	//confID := int(conf.ID)
	//return &model.SystemConfig{
	//	ID:        &confID,
	//	Key:       &conf.Key,
	//	Value:     &conf.Value,
	//	CreatedAt: &conf.CreatedAt,
	//	UpdatedAt: &conf.UpdatedAt,
	//}, nil
}

func (r *mutationResolver) AddMaterial(ctx context.Context, input model.MaterialCreateInput) (*model.Material, error) {
	return logic.AddMaterial(ctx, input)
}

func (r *mutationResolver) SaveDecodeTemplate(ctx context.Context, input model.DecodeTemplateInput) (*model.DecodeTemplate, error) {
	return logic.SaveDecodeTemplate(ctx, input)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
