package logic

import (
	"context"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
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
			Index:             parseIndexFromColumnCode(pointInput.Index),
		}
		if err := tx.Create(&point).Error; err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "point")
		}
	}

	// 为版本创建解析模板
	decodeTemplate := orm.DecodeTemplate{
		MaterialID:           version.MaterialID,
		MaterialVersionID:    version.ID,
		UserID:               user.ID,
		DataRowIndex:         15,
		CreatedAtColumnIndex: 2,
	}
	err := genDefaultProductColumns(&decodeTemplate)
	if err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeCreateObjectError, err, "material_default_decode_template")
	}
	if err := tx.Create(&decodeTemplate).Error; err != nil {
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
	if version.Active {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeActiveVersionCanNotDelete, nil)
	}

	tx := orm.Begin()
	// 删除版本
	if err := tx.Delete(&version).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "material_version")
	}
	// 删除模板
	if err := tx.Where("material_version_id = ?", version.ID).Delete(&orm.DecodeTemplate{}).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "decode_template")
	}
	// 删除点位
	if err := tx.Where("material_version_id = ?", version.ID).Delete(&orm.Point{}).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "point")
	}
	// 删除导入记录
	if err := tx.Where("material_version_id = ?", version.ID).Delete(&orm.ImportRecord{}).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "import_record")
	}
	// 删除数据
	if err := tx.Where("material_version_id = ?", version.ID).Delete(&orm.Product{}).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "products")
	}

	tx.Commit()
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

func LoadMaterialVersion(ctx context.Context, id uint) *model.MaterialVersion {
	var version orm.MaterialVersion
	if err := version.Get(id); err != nil {
		return nil
	}

	var out model.MaterialVersion
	if err := copier.Copy(&out, &version); err != nil {
		return nil
	}

	return &out
}

func ChangeMaterialVersionActive(ctx context.Context, id int, active bool) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var version orm.MaterialVersion
	if err := version.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, err.GetCode(), err, "material_version")
	}

	tx := orm.Begin()
	err := tx.Model(&orm.MaterialVersion{}).Where(
		"material_id = ? AND active = ? AND id != ?", version.MaterialID, true, version.ID,
	).Update("active", false).Error
	if err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "material_version")
	}

	version.Active = true
	if err := tx.Save(&version).Error; err != nil {
		tx.Rollback()
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "material_version")
	}
	tx.Commit()

	return model.ResponseStatusOk, nil
}
