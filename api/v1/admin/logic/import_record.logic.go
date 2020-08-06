package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
)

func ImportRecords(ctx context.Context, materialVersionID int, deviceID *int, page int, limit int, search model.ImportRecordSearch) (*model.ImportRecordsWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	query := orm.Model(&orm.ImportRecord{}).Where("material_version_id = ? AND status != ?", materialVersionID, model.ImportStatusReverted)
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}
	if search.UserID != nil {
		query = query.Where("user_id = ?", *search.UserID)
	}
	if search.Date != nil {
		query = query.Where("DATE(created_at) = DATE(?)", *search.Date)
	} else if len(search.Duration) > 0 {
		query = query.Where("DATE(created_at) >= DATE(?)", *search.Duration[0])
		if len(search.Duration) > 1 {
			query = query.Where("DATE(created_at) <= DATE(?)", *search.Duration[1])
		}
	}
	if search.FileName != nil {
		query = query.Where("file_name like ?", fmt.Sprintf("%%%s%%", *search.FileName))
	}
	if len(search.Status) > 0 {
		var status []string
		for _, s := range search.Status {
			status = append(status, string(*s))
		}

		query = query.Where("status in (?)", status)
	}

	var count int
	if err := query.Count(&count).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "import_record")
	}

	var records []orm.ImportRecord
	offset := (page - 1) * limit
	if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
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

func RevertImports(ctx context.Context, ids []int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()
	for _, id := range ids {
		if err := tx.Where("import_record_id = ?", id).Delete(&orm.Product{}).Error; err != nil {
			tx.Rollback()
			return "", errormap.SendGQLError(ctx, errormap.ErrorCodeRevertImportFailed, err)
		}

		var record orm.ImportRecord
		if err := record.Get(uint(id)); err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(ctx, err.GetCode(), err, "import_record")
		}
		if err := record.Revert(); err != nil {
			tx.Rollback()
			return "", errormap.SendGQLError(ctx, errormap.ErrorCodeRevertImportFailed, err)
		}
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func ImportData(ctx context.Context, materialID int, deviceID int, fileTokens []string) (model.ResponseStatus, error) {
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

	if err := FetchFileData(*user, material, device, fileTokens); err != nil {
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
		Yield:            record.Yield,
		Status:           model.ImportStatus(record.Status),
		RowCount:         record.RowCount,
		FileSize:         record.FileSize,
		FinishedRowCount: record.RowFinishedCount,
	}, nil
}

func ToggleBlockImports(ctx context.Context, ids []int, block bool) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	if err := orm.Model(&orm.ImportRecord{}).Where("id in (?)", ids).Update("blocked", block).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "import_record")
	}

	return model.ResponseStatusOk, nil
}
