package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
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
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	var template orm.DecodeTemplate
	if input.ID != nil {
		if err := template.Get(uint(*input.ID)); err != nil {
			return nil, errormap.SendGQLError(ctx, err.ErrorCode, err, "decode_template")
		}
	}

	template.Name = input.Name
	if template.MaterialID == 0 { // 仅创建时赋值
		template.MaterialID = uint(input.MaterialID)
	}
	if template.UserID == 0 { // 仅创建时赋值
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
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeDecodeTemplateSetDefaultFailed, err)
		}
		template.Default = *input.Default
	}

	if err := tx.Save(&template).Error; err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "decode_template")
	}
	var freshTemplate orm.DecodeTemplate
	if err := tx.Model(&freshTemplate).Where("id = ?", template.ID).First(&freshTemplate).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeObjectNotFound, err, "decode_template")
		}

		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "decode_template")
	}

	out, err := convertDecodeTemplateOutput(&freshTemplate)
	if err != nil {
		tx.Rollback()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "decode_template")
	}

	tx.Commit()
	return out, nil
}

func ListDecodeTemplate(ctx context.Context, materialID int) ([]*model.DecodeTemplate, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var templates []orm.DecodeTemplate
	if err := orm.Model(&orm.DecodeTemplate{}).Where("material_id = ?", materialID).Find(&templates).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "decode_template")
	}

	var outs []*model.DecodeTemplate
	for _, template := range templates {
		out, err := convertDecodeTemplateOutput(&template)
		if err != nil {
			continue
		}

		outs = append(outs, out)
	}

	return outs, nil
}

func convertDecodeTemplateOutput(template *orm.DecodeTemplate) (*model.DecodeTemplate, error) {
	var out model.DecodeTemplate
	if err := copier.Copy(&out, template); err != nil {
		return nil, err
	}

	var oProductColumns []*model.ProductColumn
	iProductColumns, ok := template.ProductColumns["columns"].([]interface{})
	if !ok {
		return nil, errormap.NewOrigin("type assert decode template product columns to array failed")
	}

	for _, iColumn := range iProductColumns {
		column, ok := iColumn.(map[string]interface{})
		if !ok {
			log.Warnln("type assert interface{} to map[string]interface{} failed")
			continue
		}

		var oColumn model.ProductColumn
		if name, ok := column["Name"].(string); ok {
			oColumn.Name = name
		}
		if index, ok := column["Index"].(int); ok {
			oColumn.Index = index
		}
		if cType, ok := column["Type"].(string); ok {
			oColumn.Type = model.ProductColumnType(cType)
		}

		oProductColumns = append(oProductColumns, &oColumn)
	}
	out.ProductColumns = oProductColumns
	return &out, nil
}

func DeleteDecodeTemplate(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var template orm.DecodeTemplate
	if err := template.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, err.ErrorCode, err, "decode_template")
	}

	if template.Default {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDecodeTemplateDefaultDeleteProtect, nil)
	}

	if err := orm.Delete(&template).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "decode_template")
	}

	return model.ResponseStatusOk, nil
}
