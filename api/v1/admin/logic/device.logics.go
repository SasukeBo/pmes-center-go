package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
)

func LoadDevice(ctx context.Context, deviceID uint) *model.Device {
	var device orm.Device
	if err := device.Get(deviceID); err != nil {
		return nil
	}
	var out model.Device
	if err := copier.Copy(&out, &device); err != nil {
		return nil
	}

	return &out
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

func ListDevices(ctx context.Context, pattern *string, materialID *int, page int, limit int) (*model.DeviceWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var devices []orm.Device
	var sql = orm.Model(&orm.Device{})
	if pattern != nil {
		search := fmt.Sprintf("%%%s%%", *pattern)
		sql = sql.Where("name LIKE ? OR remark LIKE ?", search, search)
	}
	if materialID != nil {
		sql = sql.Where("material_id = ?", *materialID)
	}

	var count int
	if err := sql.Count(&count).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "device")
	}

	if err := sql.Offset((page - 1) * limit).Limit(limit).Find(&devices).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "device")
	}
	var outs []*model.Device
	for _, d := range devices {
		var out model.Device
		if err := copier.Copy(&out, &d); err != nil {
			continue
		}

		outs = append(outs, &out)
	}

	return &model.DeviceWrap{
		Total:   count,
		Devices: outs,
	}, nil
}

func DeleteDevice(ctx context.Context, id int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var device orm.Device
	if err := device.Get(uint(id)); err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, err.GetCode(), err)
	}

	if err := orm.Delete(&device).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "device")
	}

	return model.ResponseStatusOk, nil
}

func Device(ctx context.Context, id int) (*model.Device, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

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
