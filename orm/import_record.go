package orm

import "github.com/jinzhu/gorm"

// 导入记录，用于记录用户从后台手动为设备导入的数据文件。
// 用于支持以文件为单位撤销数据导入。

type ImportRecord struct {
	gorm.Model
	FileName         string `gorm:"not null"`       // 文件名称
	MaterialID       int    `gorm:"not null;index"` // 关联料号ID
	DeviceID         int    `gorm:"not null;index"` // 关联设备ID
	RowCount         int    // 数据行数
	RowFinishedCount int    // 完成行数
	Finished         bool   `gorm:"not null;default:false"` // 表示处理完成
	FileSize         int
	UserID           int
	ImportType       string `gorm:"not null;default:'SYSTEM'"` // 导入方式，默认为系统
	DecodeTemplateID int    `gorm:"not null"`                   // 文件解析模板ID
}
