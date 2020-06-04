package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"github.com/jinzhu/copier"
)

func LoadDecodeTemplate(ctx context.Context, templateID uint) (*model.DecodeTemplate, error) {
	var template orm.DecodeTemplate
	if err := template.Get(templateID); err != nil {
		return nil, err
	}
	var out model.DecodeTemplate
	if err := copier.Copy(&out, &template); err != nil {
		return nil, err
	}

	return &out, nil
}

func SaveDecodeTemplate(ctx context.Context, input model.DecodeTemplateInput) (*model.DecodeTemplate, error) {
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	var template orm.DecodeTemplate
	if input.ID != nil {
		if err := template.Get(uint(*input.ID)); err != nil {
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeGetObjectFailed, err, "decode_template")
		}
	}

	template.Name = input.Name
	template.MaterialID = uint(input.MaterialID)
	if template.UserID == 0 {
		template.UserID = user.ID
	}
	if input.Description != nil {
		template.Description = *input.Description
	}
	template.DataRowIndex = input.DataRowIndex
	template.CreatedAtColumnIndex = input.CreatedAtColumnIndex

	var productColumns []orm.Column
	for _, iColumn := range input.ProductColumns {
		var column orm.Column
		if err := copier.Copy(&column, &iColumn); err != nil {
			continue
		}
		productColumns = append(productColumns, column)
	}
	template.ProductColumns = types.Map{"columns": productColumns}
	template.PointColumns = input.PointColumns
	if input.Default != nil {
		if err := tx.Model(&orm.DecodeTemplate{}).Where("material_id = ? AND decode_templates.default = ?", input.MaterialID, true).Update("default", false).Error; err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeDecodeTemplateSetDefaultFailed, err)
		}
		template.Default = *input.Default
	}

	if err := tx.Save(&template).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeSaveObjectError, err, "decode_template")
	}

	var out model.DecodeTemplate
	if err := copier.Copy(&out, &template); err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "decode_template")
	}

	var oProductColumns []*model.ProductColumn
	for _, column := range template.ProductColumns {
		var oColumn model.ProductColumn
		if err := copier.Copy(&oColumn, &column); err != nil {
			continue
		}

		oProductColumns = append(oProductColumns, &oColumn)
	}
	out.ProductColumns = oProductColumns

	tx.Commit()
	return &out, nil
}
