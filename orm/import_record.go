package orm

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/jinzhu/gorm"
)

// 导入记录，用于记录用户从后台手动为设备导入的数据文件。
// 用于支持以文件为单位撤销数据导入。

type ImportStatus string

const (
	ImportRecordTypeSystem = "SYSTEM"
	ImportRecordTypeUser   = "USER"

	ImportStatusLoading  ImportStatus = "Loading"
	ImportStatusFinished ImportStatus = "Finished"
	ImportStatusFailed   ImportStatus = "Failed"
	ImportStatusReverted ImportStatus = "Reverted"
)

type ImportRecord struct {
	gorm.Model
	FileID             uint          `gorm:"column:file_id"'` // 关联文件的ID
	FileName           string       `gorm:"not null"`        // 文件名称
	Path               string       `gorm:"not null"`        // 存储路径
	MaterialID         uint         `gorm:"not null;index"`  // 关联料号ID
	DeviceID           uint         `gorm:"not null;index"`  // 关联设备ID
	RowCount           int          // 数据行数
	RowFinishedCount   int          // 完成行数
	Status             ImportStatus `gorm:"not null;default:false"` // 导入状态
	ErrorCode          string       // 错误码
	OriginErrorMessage string       // 原始错误信息
	FileSize           int
	UserID             uint
	ImportType         string `gorm:"not null;default:'SYSTEM'"` // 导入方式，默认为系统
	DecodeTemplateID   uint   `gorm:"not null"`                  // 文件解析模板ID
}

func (i *ImportRecord) Get(id uint) *errormap.Error {
	if err := DB.Model(i).Where("id = ?", id).First(i).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (i *ImportRecord) Finish() error {
	i.Status = ImportStatusFinished
	return Save(i).Error
}

func (i *ImportRecord) Failed(errorCode string, origin interface{}) error {
	i.Status = ImportStatusFailed
	i.ErrorCode = errorCode
	i.OriginErrorMessage = fmt.Sprint(origin)
	return Save(i).Error
}

func (i *ImportRecord) Revert() error {
	i.Status = ImportStatusReverted
	return Save(i).Error
}

func (i *ImportRecord) Load() error {
	i.Status = ImportStatusLoading
	return Save(i).Error
}
