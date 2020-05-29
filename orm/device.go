package orm

// 料号的生产设备
// 生产设备的创建方式有两种
// 1.通过数据文件名称解析
// 2.通过后台手动创建

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Device struct {
	gorm.Model
	UUID           string `gorm:"column:uuid;unique_index;not null"`
	Name           string `gorm:"not null;unique_index"`
	IP             string `gorm:"column:ip;"`
	MaterialID     int    `gorm:"column:material_id;not null;index"`
	DeviceSupplier string
	IsRealtime     bool `gorm:"default:false;not null"`
	Address        string
}

func (d *Device) BeforeCreate() error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	d.UUID = uid.String()
	return nil
}
