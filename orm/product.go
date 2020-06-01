package orm

import (
	"github.com/SasukeBo/ftpviewer/orm/types"
	"time"
)

// Product 产品表
type Product struct {
	ID             int       `gorm:"column:id;primary_key"`
	ImportRecordID int       `gorm:"column:import_record_id;not null;index"`
	MaterialID     int       `gorm:"column:material_id;not null;index"`
	DeviceID       int       `gorm:"column:device_id;not null;index"`
	Qualified      bool      `gorm:"column:qualified;default:false"`
	CreatedAt      time.Time `gorm:"index"`
	Attribute      types.Map `type:JSON;not null`
	PointValues    types.Map `gorm:"type:JSON;not null"`
}
