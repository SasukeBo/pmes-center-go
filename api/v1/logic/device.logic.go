package logic

import (
	"context"
	"errors"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
	"time"
)

func Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	var devices []orm.Device
	if err := orm.DB.Where("material_id = ?", materialID).Find(&devices).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "device")
	}
	var outs []*model.Device
	for _, device := range devices {
		var out model.Device
		if err := copier.Copy(&out, device); err != nil {
			log.Errorln(err)
			continue
		}
		outs = append(outs, &out)
	}
	return outs, nil
}

func AnalyzeDevices(ctx context.Context, materialID int, versionID *int) ([]*model.DeviceResult, error) {
	var version orm.MaterialVersion
	if versionID != nil {
		if err := version.Get(uint(*versionID)); err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
		}
	} else {
		err := orm.Model(&orm.MaterialVersion{}).Where("material_id = ? AND active = ?", materialID, true).Find(&version).Error
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
		}
	}

	var devices []orm.Device
	if err := orm.DB.Model(&orm.Device{}).Where("material_id = ?", materialID).Find(&devices).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "device")
	}

	var outs []*model.DeviceResult
	for _, d := range devices {
		var out = model.DeviceResult{
			Device: &model.Device{
				ID:   int(d.ID),
				Name: d.Name,
			},
		}
		query := orm.Model(&orm.Product{}).Select("COUNT(products.id), products.qualified")
		query = query.Joins("JOIN import_records ON products.import_record_id = import_records.id")
		query = query.Where("products.device_id = ? AND import_records.blocked = ?", d.ID, false)
		query = query.Where("products.material_version_id = ?", version.ID)
		rows, err := query.Group("products.qualified").Rows()
		if err != nil {
			outs = append(outs, &out)
			continue
		}

		for rows.Next() {
			var amount int
			var qualified bool
			if err := rows.Scan(&amount, &qualified); err != nil {
				continue
			}
			if qualified {
				out.Ok = amount
			} else {
				out.Ng = amount
			}
		}
		outs = append(outs, &out)
	}

	return outs, nil
}

func AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	if searchInput.DeviceID == nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("missing device_id"))
	}
	var device orm.Device
	if err := device.Get(uint(*searchInput.DeviceID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "device")
	}
	var material orm.Material
	if err := material.Get(device.MaterialID); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	beginTime := searchInput.BeginTime
	endTime := searchInput.EndTime
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}
	if beginTime == nil {
		t := endTime.AddDate(-1, 0, 0)
		beginTime = &t
	}

	out := model.Device{
		ID:   int(device.ID),
		Name: device.Name,
	}

	var ok int
	var ng int
	query := orm.Model(&orm.Product{}).Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("import_records.blocked = ? AND products.device_id = ?", false, searchInput.DeviceID)
	query = query.Where("products.created_at < ? AND products.created_at > ?", endTime, beginTime)
	query.Where("products.qualified = ?", true).Count(&ok)
	query.Where("products.qualified = ?", false).Count(&ng)

	return &model.DeviceResult{
		Device: &out,
		Ok:     ok,
		Ng:     ng,
	}, nil
}

func GroupAnalyzeDevice(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return groupAnalyze(ctx, analyzeInput, "device", nil)
}

func Device(ctx context.Context, id int) (*model.Device, error) {
	var device orm.Device
	if err := device.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "device")
	}

	var out model.Device
	if err := copier.Copy(&out, &device); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "device")
	}

	return &out, nil
}
