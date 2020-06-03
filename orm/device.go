package orm

// 料号的生产设备
// 生产设备的创建方式有两种
// 1.通过数据文件名称解析
// 2.通过后台手动创建

import (
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type Device struct {
	gorm.Model
	UUID           string `gorm:"column:uuid;unique_index;not null"`
	Name           string `gorm:"not null"`                                    // 用于存储用户指定的设备名称，不指定时，默认为Remark的值
	Remark         string `gorm:"not null;unique_index:uidx_name_material_id"` // 用于存储从数据文件解析出的名称
	IP             string `gorm:"column:ip;"`
	MaterialID     uint   `gorm:"column:material_id;not null;unique_index:uidx_name_material_id"` // 同一料号下的设备remark不可重复
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

func (d *Device) GetWithName(name string) *errormap.Error {
	if err := DB.Model(d).Where("name = ?", name).First(d).Error; err != nil {
		return handleError(err, "name", name)
	}

	return nil
}

func (d *Device) Get(id uint) *errormap.Error {
	if err := DB.Model(d).Where("id = ?", id).First(d).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (d *Device) CreateIfNotExist(materialID uint, remark string) error {
	DB.Model(d).Where("material_id = ? AND remark = ?", materialID, remark).First(d)
	if d.ID == 0 {
		d.Name = remark
		d.MaterialID = materialID
		d.Remark = remark
		err := DB.Create(d).Error
		return err
	}

	return nil
}
