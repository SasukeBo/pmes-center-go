package orm

import (
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/jinzhu/gorm"
)

// 导入记录，用于记录用户从后台手动为设备导入的数据文件。
// 用于支持以文件为单位撤销数据导入。

const (
	ImportRecordTypeSystem = "SYSTEM"
	ImportRecordTypeUser   = "USER"
)

type ImportRecord struct {
	gorm.Model
	FileName         string `gorm:"not null"`       // 文件名称
	Path             string `gorm:"not null"`       // 存储路径
	MaterialID       uint   `gorm:"not null;index"` // 关联料号ID
	DeviceID         uint   `gorm:"not null;index"` // 关联设备ID
	RowCount         int    // 数据行数
	RowFinishedCount int    // 完成行数
	Finished         bool   `gorm:"not null;default:false"` // 表示处理完成
	Error            string // 错误信息
	FileSize         int
	UserID           uint
	ImportType       string `gorm:"not null;default:'SYSTEM'"` // 导入方式，默认为系统
	DecodeTemplateID uint   `gorm:"not null"`                  // 文件解析模板ID
}

func (i *ImportRecord) Get(id uint) *errormap.Error {
	if err := DB.Model(i).Where("id = ?", id).First(i).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}
