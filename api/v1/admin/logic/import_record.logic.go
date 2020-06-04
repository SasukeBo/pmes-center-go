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
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	sql := orm.Model(&orm.ImportRecord{}).Where("material_id = ?", materialID)
	if deviceID != nil {
		sql = sql.Where("device_id = ?", *deviceID)
	}
	var count int
	if err := sql.Count(&count).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCountObjectFailed, err, "import_record")
	}

	var records []orm.ImportRecord
	offset := (page - 1) * limit
	if err := sql.Order("created_at desc").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeGetObjectFailed, err, "import_record")
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
