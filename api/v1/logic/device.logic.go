package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
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
