package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
	"reflect"
	"strconv"
)

func ProductScrollFetch(ctx context.Context, searchInput model.ProductSearch, limit int, offset int) (*model.ProductWrap, error) {
	var material orm.Material
	if err := material.Get(uint(searchInput.MaterialID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	if offset < 0 || limit < 0 {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errormap.NewOrigin("limit and offset cannot little than 0: limit=%v offset=%v", limit, offset))
	}

	sql := orm.Model(&orm.Product{}).Where("material_id = ?", material.ID)
	if searchInput.End != nil {
		sql = sql.Where("created_at < ?", *searchInput.End)
	}
	if searchInput.Begin != nil {
		sql = sql.Where("created_at > ?", *searchInput.Begin)
	}
	if searchInput.DeviceID != nil {
		var device orm.Device
		if err := device.Get(uint(*searchInput.DeviceID)); err == nil {
			sql = sql.Where("device_id = ?", device.ID)
		}
	}
	if searchInput.ImportRecordID != nil {
		var importRecord orm.ImportRecord
		if err := importRecord.Get(uint(*searchInput.ImportRecordID)); err == nil {
			sql = sql.Where("import_record_id = ?", importRecord.ID)
		}
	}
	if searchInput.Qualified != nil {
		sql = sql.Where("qualified = ?", *searchInput.Qualified)
	}

	for k, v := range searchInput.Attributes {
		tv := reflect.TypeOf(v)
		if tv.Name() == "Number" {
			number, err := strconv.ParseFloat(fmt.Sprint(v), 64)
			if err != nil {
				number = 0
			}
			sql = sql.Where("JSON_EXTRACT(`attribute`, ?) = ?", fmt.Sprintf("$.\"%s\"", k), number)
			continue
		}

		sql = sql.Where("JSON_EXTRACT(`attribute`, ?) = ?", fmt.Sprintf("$.\"%s\"", k), v)
	}

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "product")
	}

	var products []orm.Product
	if err := sql.Order("id asc").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "product")
	}

	var outData []*model.Product
	for _, p := range products {
		var data model.Product
		if err := copier.Copy(&data, &p); err == nil {
			outData = append(outData, &data)
		}
	}

	return &model.ProductWrap{
		Data:  outData,
		Total: total,
	}, nil
}
