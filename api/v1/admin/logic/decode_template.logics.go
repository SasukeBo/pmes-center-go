package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"math"
	"strings"
)

func LoadDecodeTemplate(ctx context.Context, templateID uint) *model.DecodeTemplate {
	var template orm.DecodeTemplate
	if err := template.Get(templateID); err != nil {
		return nil
	}
	var out model.DecodeTemplate
	if err := copier.Copy(&out, &template); err != nil {
		return nil
	}

	return &out
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
	template.CreatedAtColumnIndex = parseIndexFromColumnCode(input.CreatedAtColumnIndex)
	log.Warn("parse createdAtColumnIndex from %v to %v", input.CreatedAtColumnIndex, template.CreatedAtColumnIndex)

	productColumns := make(types.Map)
	for _, iColumn := range input.ProductColumns {
		var column orm.Column
		if err := copier.Copy(&column, &iColumn); err != nil {
			continue
		}
		column.Index = parseIndexFromColumnCode(iColumn.Index)
		productColumns[iColumn.Token] = column
	}
	template.ProductColumns = productColumns
	pointColumns := make(types.Map)

	for k, v := range input.PointColumns {
		if code, ok := v.(string); ok {
			pointColumns[k] = parseIndexFromColumnCode(code)
		}
	}

	template.PointColumns = pointColumns
	template.Default = input.Default
	if template.Default {
		if err := tx.Model(&orm.DecodeTemplate{}).Where("material_id = ? AND decode_templates.default = ?", input.MaterialID, true).Update("default", false).Error; err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeDecodeTemplateSetDefaultFailed, err)
		}
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

	out.CreatedAtColumnIndex = parseColumnCodeFromIndex(template.CreatedAtColumnIndex)
	var oProductColumns []*model.ProductColumn
	iProductColumns := template.ProductColumns

	for token, iColumn := range iProductColumns {
		column, ok := iColumn.(map[string]interface{})
		if !ok {
			log.Warnln("type assert interface{} to map[string]interface{} failed")
			continue
		}

		var oColumn model.ProductColumn
		oColumn.Token = token
		if label, ok := column["Label"].(string); ok {
			oColumn.Label = label
		}

		if idx, ok := column["Index"].(float64); ok {
			oColumn.Index = parseColumnCodeFromIndex(int(idx))
		}

		if cType, ok := column["Type"].(string); ok {
			oColumn.Type = model.ProductColumnType(cType)
		}

		if prefix, ok := column["Prefix"].(string); ok {
			oColumn.Prefix = prefix
		}

		oProductColumns = append(oProductColumns, &oColumn)
	}
	out.ProductColumns = oProductColumns

	oPointColumns := make(map[string]interface{})
	for k, v := range template.PointColumns {
		if idx, ok := v.(float64); ok {
			oPointColumns[k] = parseColumnCodeFromIndex(int(idx))
		}
	}

	out.PointColumns = oPointColumns
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

// 存储列号时，将用户输入的英文列号转换为数组index，从0开始
func parseIndexFromColumnCode(columnCode string) int {
	rs := []rune(columnCode)
	length := len(rs)
	var sum int
	for i, r := range rs {
		ascii := int(r)
		sum = sum + int(math.Pow(26, float64(length-1-i)))*(ascii-64)
	}
	return sum
}

const charAASCII = 65

// 读取存储列号时，将数组index加1解析为英文列号
func parseColumnCodeFromIndex(index int) string {
	offsets := make([]int, 0)
	for {
		if index <= 26 {
			offsets = append(offsets, index)
			break
		}

		offset := index % 26
		if offset == 0 {
			offset = 26
			index = index/26 - 1
		} else {
			index = index / 26
		}
		offsets = append(offsets, offset)
	}

	codes := make([]string, 0)
	length := len(offsets)
	for i := length - 1; i >= 0; i-- {
		ascii := charAASCII - 1 + offsets[i]
		code := string(ascii)
		codes = append(codes, code)
	}

	return strings.Join(codes, "")
}

func ChangeDefaultTemplate(ctx context.Context, id int, isDefault bool) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	var template orm.DecodeTemplate
	if err := template.Get(uint(id)); err != nil {
		tx.Rollback()
		return "", errormap.SendGQLError(ctx, err.GetCode(), err, "decode_template")
	}

	if isDefault {
		err := tx.Model(&orm.DecodeTemplate{}).Where(
			"material_id = ? AND decode_templates.default = ?",
			template.MaterialID, true,
		).Update("default", false).Error
		if err != nil {
			tx.Rollback()
			return "", errormap.SendGQLError(ctx, errormap.ErrorCodeDecodeTemplateSetDefaultFailed, err)
		}
	}

	if err := tx.Model(&orm.DecodeTemplate{}).Where("id = ?", id).Update("default", isDefault).Error; err != nil {
		tx.Rollback()
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "decode_template")
	}
	tx.Commit()

	return model.ResponseStatusOk, nil
}

func genDefaultProductColumns(template *orm.DecodeTemplate) error {
	productColumns := make(types.Map)
	var config orm.SystemConfig
	err := config.GetConfig(orm.SystemConfigProductColumnHeadersKey)
	if err != nil {
		return err
	}

	headers := strings.Split(config.Value, ";")
	fmt.Println("[headers]", headers)
	for _, header := range headers {
		vs := strings.Split(header, ":")
		if len(vs) < 4 {
			continue
		}

		token := vs[1]
		prefix := strings.ToUpper(token[:1])
		var column = orm.Column{
			Index:  parseIndexFromColumnCode(vs[0]),
			Token:  token,
			Label:  vs[2],
			Type:   vs[3],
			Prefix: prefix,
		}

		productColumns[token] = column
	}

	template.ProductColumns = productColumns
	return nil
}
