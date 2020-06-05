package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
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

func SaveDevice(ctx context.Context, input model.DeviceInput) (*model.Device, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var device orm.Device
	if input.ID != nil {
		if err := device.Get(uint(*input.ID)); err != nil {
			return nil, errormap.SendGQLError(ctx, err.ErrorCode, err, "device")
		}
	}
	device.Name = input.Name
	if device.ID == 0 {
		device.Remark = input.Remark
		device.MaterialID = uint(input.MaterialID)
	}
	if input.IP != nil {
		device.IP = *input.IP
	}
	if input.Address != nil {
		device.Address = *input.Address
	}
	if input.DeviceSupplier != nil {
		device.DeviceSupplier = *input.DeviceSupplier
	}
	device.IsRealtime = input.IsRealtime

	if err := orm.Save(&device).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "device")
	}

	var out model.Device
	if err := copier.Copy(&out, &device); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "device")
	}

	return &out, nil
}
