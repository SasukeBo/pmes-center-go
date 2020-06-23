package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
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

func ImportData(ctx context.Context, materialID int, deviceID int, decodeTemplateID int, fileTokens []*string) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return "", errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	// TODO: finish it
	return "", nil
}
