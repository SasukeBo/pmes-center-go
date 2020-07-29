package orm

import (
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"time"
)

// Product 产品表
type Product struct {
	ID                uint      `gorm:"column:id;primary_key"`
	ImportRecordID    uint      `gorm:"column:import_record_id;not null;index"`
	MaterialVersionID uint      `gorm:"index"`
	MaterialID        uint      `gorm:"column:material_id;not null;index"`
	DeviceID          uint      `gorm:"column:device_id;not null;index"`
	Qualified         bool      `gorm:"column:qualified;default:false"`
	BarCode           string    `gorm:"column:bar_code;"`
	CreatedAt         time.Time `gorm:"index"` // 检测时间
	Attribute         types.Map `gorm:"type:JSON;not null"`
	PointValues       types.Map `gorm:"type:JSON;not null"`
}
