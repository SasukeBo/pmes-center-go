package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
)

// AddMaterial 创建料号
// 创建料号执行以下操作：
// - 创建料号记录
// - 创建料号版本记录
// - 为料号版本创建解析模板
// - 为料号版本创建检测尺寸
func AddMaterial(ctx context.Context, input model.MaterialCreateInput) (*model.Material, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	// 创建料号
	var material orm.Material
	tx.Model(&material).Where("name = ?", input.Name).First(&material)
	if material.ID != 0 {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAlreadyExists, nil)
	}
	material.Name = input.Name
	if input.CustomerCode != nil {
		material.CustomerCode = *input.CustomerCode
	}
	if input.ProjectRemark != nil {
		material.ProjectRemark = *input.ProjectRemark
	}
	if input.YieldScore != nil {
		material.YieldScore = *input.YieldScore
	}
	if err := tx.Create(&material).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material")
	}

	// 创建料号版本
	var materialVersion = orm.MaterialVersion{
		Version:     "Init Version",
		Description: "初始化版本",
		MaterialID:  material.ID,
		UserID:      user.ID,
		Active:      true,
	}
	if err := tx.Create(&materialVersion).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material")
	}

	// 为版本创建默认解析模板
	decodeTemplate := orm.DecodeTemplate{
		Name:                 "默认模板",
		MaterialID:           material.ID,
		MaterialVersionID:    materialVersion.ID,
		UserID:               user.ID,
		Description:          "创建料号时自动生成的默认解析模板",
		DataRowIndex:         15,
		CreatedAtColumnIndex: 2,
		Default:              true,
	}
	err := genDefaultProductColumns(&decodeTemplate)
	if err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material_default_decode_template")
	}

	pointColumns := make(types.Map)
	pointStartIndex := parseIndexFromColumnCode(configer.GetString("default_point_begin_index"))
	for i, pointInput := range input.Points {
		point := orm.Point{
			Name:              pointInput.Name,
			MaterialID:        material.ID,
			MaterialVersionID: materialVersion.ID,
			UpperLimit:        pointInput.UpperLimit,
			LowerLimit:        pointInput.LowerLimit,
			Nominal:           pointInput.Nominal,
		}
		if err := tx.Create(&point).Error; err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "point")
		}
		pointColumns[point.Name] = i + pointStartIndex
	}
	decodeTemplate.PointColumns = pointColumns

	if err := tx.Create(&decodeTemplate).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material_default_decode_template")
	}

	tx.Commit()

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}

	return &out, nil
}

func Materials(ctx context.Context, pattern *string, page int, limit int) (*model.MaterialWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	sql := orm.Model(&orm.Material{})
	if pattern != nil {
		search := fmt.Sprintf("%%%s%%", *pattern)
		sql = sql.Where("name LIKE ? OR customer_code LIKE ? OR project_remark LIKE ?", search, search, search)
	}

	var materials []orm.Material
	offset := (page - 1) * limit
	if err := sql.Order("id desc").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material")
	}

	var outs []*model.Material
	for _, i := range materials {
		var out model.Material
		if err := copier.Copy(&out, &i); err != nil {
			continue
		}

		outs = append(outs, &out)
	}

	var count int
	if err := sql.Model(&orm.Material{}).Count(&count).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "material")
	}
	return &model.MaterialWrap{
		Total:     count,
		Materials: outs,
	}, nil
}

func LoadMaterial(ctx context.Context, materialID uint) *model.Material {
	var material orm.Material
	if err := material.Get(materialID); err != nil {
		return nil
	}
	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil
	}

	return &out
}

func DeleteMaterial(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	var material orm.Material
	if err := material.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, err.ErrorCode, err, "material")
	}

	if err := tx.Delete(orm.Device{}, "material_id = ?", material.ID).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material_devices")
	}

	if err := tx.Delete(orm.ImportRecord{}, "material_id = ?", material.ID).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material_import_records")
	}

	if err := tx.Delete(orm.Product{}, "material_id = ?", material.ID).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "products")
	}

	if err := tx.Delete(orm.DecodeTemplate{}, "material_id = ?", material.ID).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material_decode_templates")
	}

	if err := tx.Delete(&material).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material")
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func UpdateMaterial(ctx context.Context, input model.MaterialUpdateInput) (*model.Material, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var material orm.Material
	if err := material.Get(uint(input.ID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.ErrorCode, err, "material")
	}

	if input.ProjectRemark != nil {
		material.ProjectRemark = *input.ProjectRemark
	}
	if input.CustomerCode != nil {
		material.CustomerCode = *input.CustomerCode
	}
	if input.YieldScore != nil {
		material.YieldScore = *input.YieldScore
	}
	if err := orm.Save(&material).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "material")
	}
	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}
	return &out, nil
}

func Material(ctx context.Context, id int) (*model.Material, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var material orm.Material
	if err := material.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}

	return &out, nil
}

func MaterialFetch(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var material orm.Material
	if err := material.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	// 解析FTP服务器指定料号路径下的所有未解析文件
	if err := FetchMaterialData(&material); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialDataFetchFailed, err)
	}

	return model.ResponseStatusOk, nil
}
