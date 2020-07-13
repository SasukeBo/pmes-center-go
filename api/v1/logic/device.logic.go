package logic

import (
	"context"
	"errors"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/log"
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

func AnalyzeDevices(ctx context.Context, materialID int) ([]*model.DeviceResult, error) {
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
		rows, err := orm.DB.Model(&orm.Product{}).Where("device_id = ?", d.ID).Select("COUNT(id), qualified").Group("qualified").Rows()
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
	orm.DB.Model(&orm.Product{}).Where(
		"device_id = ? and created_at < ? and created_at > ? and qualified = 1",
		searchInput.DeviceID, endTime, beginTime,
	).Count(&ok)
	orm.DB.Model(&orm.Product{}).Where(
		"device_id = ? and created_at < ? and created_at > ? and qualified = 0",
		searchInput.DeviceID, endTime, beginTime,
	).Count(&ng)

	return &model.DeviceResult{
		Device: &out,
		Ok:     ok,
		Ng:     ng,
	}, nil
}

func GroupAnalyzeDevice(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return groupAnalyze(ctx, analyzeInput, "device")
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
