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
	D2Code         string    `gorm:"column:d2_code"`
	LineID         string    `gorm:"column:line_id;index"`
	JigID          string    `gorm:"column:jig_id;index"`
	MouldID        string    `gorm:"column:mould_id;index"`
	ShiftNumber    string    `gorm:"index"`
	PointValues    types.Map `gorm:"type:JSON;not null"`
}
