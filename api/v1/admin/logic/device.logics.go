package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
)

func LoadDevice(ctx context.Context, deviceID uint) (*model.Device, error) {
	var device orm.Device
	if err := device.Get(deviceID); err != nil {
		return nil, err
	}
	var out model.Device
	if err := copier.Copy(&out, &device); err != nil {
		return nil, err
	}

	return &out, nil
}
