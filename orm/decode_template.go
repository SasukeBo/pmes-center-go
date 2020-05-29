package orm

import (
	"github.com/SasukeBo/ftpviewer/orm/types"
	"github.com/jinzhu/gorm"
)

// 数据文件解析模板
// 用于指定文件的必要数据位置

type DecodeTemplate struct {
	gorm.Model
	Name           string `gorm:"not null"`
	MaterialID     int    `gorm:"not null"`
	UserID         int
	Description    string
	DataRowIndex   int
	ProductColumns types.Map `gorm:"type:JSON;not null"`
	PointColumns   types.Map `gorm:"type:JSON;not null"`
}
