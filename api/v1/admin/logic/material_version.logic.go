package logic

import (
	"context"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
)

// CreateMaterialVersion 新建料号版本
// - 新增版本
// - 新增版本的检测尺寸
// - 新增版本的解析模板
func CreateMaterialVersion(ctx context.Context, input model.MaterialVersionInput) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	// 创建版本
	var version = orm.MaterialVersion{
		Version:    input.Version,
		MaterialID: uint(input.MaterialID),
		UserID:     user.ID,
	}

	if input.Description != nil {
		version.Description = *input.Description
	}
	if input.Active != nil {
		version.Active = *input.Active
	}
	if err := tx.Create(&version).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material_version")
	}

	// 新增版本检测尺寸
	for _, pointInput := range input.Points {
		point := orm.Point{
			Name:              pointInput.Name,
			MaterialID:        uint(input.MaterialID),
			MaterialVersionID: version.ID,
			UpperLimit:        pointInput.UpperLimit,
			LowerLimit:        pointInput.LowerLimit,
			Nominal:           pointInput.Nominal,
		}
		if err := tx.Create(&point).Error; err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "point")
		}
	}

	// 创建解析模板
	var templateInput = input.Template
	var template = orm.DecodeTemplate{
		Name:                 templateInput.Name,
		MaterialID:           uint(input.MaterialID),
		MaterialVersionID:    version.ID,
		UserID:               user.ID,
		DataRowIndex:         templateInput.DataRowIndex,
		CreatedAtColumnIndex: parseIndexFromColumnCode(templateInput.CreatedAtColumnIndex),
		Default:              true,
	}

	pointColumns := make(types.Map)
	for k, v := range templateInput.PointColumns {
		if code, ok := v.(string); ok {
			pointColumns[k] = parseIndexFromColumnCode(code)
		}
	}
	template.PointColumns = pointColumns

	productColumns := make(types.Map)
	for _, iColumn := range templateInput.ProductColumns {
		var column orm.Column
		if err := copier.Copy(&column, &iColumn); err != nil {
			continue
		}
		column.Index = parseIndexFromColumnCode(iColumn.Index)
		productColumns[iColumn.Token] = column
	}
	template.ProductColumns = productColumns

	if templateInput.Description != nil {
		template.Description = *templateInput.Description
	}

	if err := tx.Create(&template).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material_decode_template")
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func DeleteMaterialVersion(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var version orm.MaterialVersion
	if err := version.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
	}

	if err := orm.Delete(&version).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material_version")
	}

	return model.ResponseStatusOk, nil
}

func UpdateMaterialVersion(ctx context.Context, id int, input model.MaterialVersionUpdateInput) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var version orm.MaterialVersion
	if err := version.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
	}

	if input.Description != nil {
		version.Description = *input.Description
	}
	if input.Version != nil {
		version.Description = *input.Version
	}
	if input.Active != nil {
		version.Active = *input.Active
	}

	if err := orm.Save(&version).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "material_version")
	}

	return model.ResponseStatusOk, nil
}

func MaterialVersions(ctx context.Context, id int) ([]*model.MaterialVersion, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var versions []orm.MaterialVersion
	if err := orm.Model(&orm.MaterialVersion{}).Where("material_id = ?", id).Find(&versions).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
	}

	var outs []*model.MaterialVersion
	for _, v := range versions {
		var out model.MaterialVersion
		if err := copier.Copy(&out, &v); err != nil {
			log.Errorln(err)
			continue
		}

		outs = append(outs, &out)
	}

	return outs, nil
}
