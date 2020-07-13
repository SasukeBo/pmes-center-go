package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
)

func ImportRecords(ctx context.Context, materialID int, deviceID *int, page int, limit int) (*model.ImportRecordsWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	sql := orm.Model(&orm.ImportRecord{}).Where("material_id = ?", materialID)
	if deviceID != nil {
		sql = sql.Where("device_id = ?", *deviceID)
	}
	var count int
	if err := sql.Count(&count).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "import_record")
	}

	var records []orm.ImportRecord
	offset := (page - 1) * limit
	if err := sql.Order("created_at desc").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "import_record")
	}

	var outs []*model.ImportRecord
	for _, r := range records {
		var out model.ImportRecord
		if err := copier.Copy(&out, &r); err != nil {
			continue
		}
		outs = append(outs, &out)
	}

	return &model.ImportRecordsWrap{
		Total:         count,
		ImportRecords: outs,
	}, nil
}

func RevertImport(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var record orm.ImportRecord
	if err := record.Get(uint(id)); err != nil {
		return "", errormap.SendGQLError(ctx, err.GetCode(), err, "import_record")
	}

	tx := orm.Begin()
	if err := tx.Where("import_record_id = ?", record.ID).Delete(&orm.Product{}).Error; err != nil {
		tx.Rollback()
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodeRevertImportFailed, err)
	}

	err := tx.Model(&orm.ImportRecord{}).Where("id = ?", record.ID).Update("status", orm.ImportStatusReverted).Error
	if err != nil {
		tx.Rollback()
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodeRevertImportFailed, err)
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func ImportData(ctx context.Context, materialID int, deviceID int, decodeTemplateID int, fileTokens []string) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var material orm.Material
	if err := material.Get(uint(materialID)); err != nil {
		return "", errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	var device orm.Device
	if err := device.Get(uint(deviceID)); err != nil {
		return "", errormap.SendGQLError(ctx, err.GetCode(), err, "device")
	}

	var template orm.DecodeTemplate
	if err := template.Get(uint(decodeTemplateID)); err != nil {
		return "", errormap.SendGQLError(ctx, err.GetCode(), err, "decode_template")
	}

	if err := FetchFileData(*user, material, device, template, fileTokens); err != nil {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodeImportFailedWithPanic, err)
	}

	return model.ResponseStatusOk, nil
}

func MyImportRecords(ctx context.Context, page int, limit int) (*model.ImportRecordsWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	sql := orm.Model(&orm.ImportRecord{}).Where("user_id = ?", user.ID).Order("id desc")

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "import_record")
	}

	var records []orm.ImportRecord
	if err := sql.Offset((page - 1) * limit).Limit(limit).Find(&records).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "import_record")
	}

	var outs []*model.ImportRecord
	for _, r := range records {
		var out model.ImportRecord
		if err := copier.Copy(&out, &r); err != nil {
			log.Error("Copy object failed for record(id=%v): %v", r.ID, err)
			continue
		}

		outs = append(outs, &out)
	}

	return &model.ImportRecordsWrap{
		Total:         total,
		ImportRecords: outs,
	}, nil
}

func ImportStatus(ctx context.Context, id int) (*model.ImportStatusResponse, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var record orm.ImportRecord
	if err := record.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "import_record")
	}

	return &model.ImportStatusResponse{
		Status:           model.ImportStatus(record.Status),
		FinishedRowCount: record.RowFinishedCount,
	}, nil
}
