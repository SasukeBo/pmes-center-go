package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"time"
)

func (r *queryResolver) AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	if searchInput.DeviceID == nil {
		return nil, NewGQLError("设备ID不能为空", "searchInput.DeviceID can't be empty")
	}
	device := orm.GetDeviceWithID(*searchInput.DeviceID)
	if device == nil {
		return nil, NewGQLError("设备不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.DeviceID))
	}
	material := orm.GetMaterialWithID(device.MaterialID)
	if material == nil {
		return nil, NewGQLError("设备生产的料号不存在", fmt.Sprintf("get material with id = %v failed", device.MaterialID))
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
		ID:   &device.ID,
		Name: &device.Name,
	}

	// TODO: 关闭自动拉取
	//fileIDs, err := logic.NeedFetch(material, beginTime, endTime)
	//if err != nil {
	//	return nil, err
	//}
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内设备数据")}
	//	return &model.DeviceResult{Status: status, Device: &out}, nil
	//}

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
		Ok:     &ok,
		Ng:     &ng,
	}, nil
}

func (r *queryResolver) Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	var devices []orm.Device
	if err := orm.DB.Where("material_id = ?", materialID).Find(&devices).Error; err != nil {
		return nil, NewGQLError("获取设备信息失败", err.Error())
	}
	var outs []*model.Device
	for _, i := range devices {
		v := i
		outs = append(outs, &model.Device{
			ID:   &v.ID,
			Name: &v.Name,
		})
	}
	return outs, nil
}
