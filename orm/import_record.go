package orm

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"time"
)

// 导入记录，用于记录用户从后台手动为设备导入的数据文件。
// 用于支持以文件为单位撤销数据导入。

type ImportStatus string

const (
	ImportRecordTypeSystem   = "SYSTEM"
	ImportRecordTypeRealtime = "REALTIME"
	ImportRecordTypeUser     = "USER"

	ImportStatusLoading   ImportStatus = "Loading"
	ImportStatusImporting ImportStatus = "Importing"
	ImportStatusFinished  ImportStatus = "Finished"
	ImportStatusFailed    ImportStatus = "Failed"
	ImportStatusReverted  ImportStatus = "Reverted"
)

type ImportRecord struct {
	gorm.Model
	FileID             uint         `gorm:"column:file_id"'` // 关联文件的ID
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
	ImportType         string  `gorm:"not null;default:'SYSTEM'"` // 导入方式，默认为系统
	DecodeTemplateID   uint    `gorm:"not null"`                  // 文件解析模板ID
	MaterialVersionID  uint    `gorm:"not null;default: 0"`       // 料号版本信息
	Blocked            bool    `gorm:"default:false"`             // 屏蔽导入的数据
	Yield              float64 // 单次导入记录的良率
}

func (i *ImportRecord) Get(id uint) *errormap.Error {
	if err := DB.Model(i).Where("id = ?", id).First(i).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (i *ImportRecord) genKey(deviceID uint) string {
	return fmt.Sprintf("device_realtime_key_%v_%s", deviceID, util.NowDateStr())
}

func (i *ImportRecord) Finish(yield float64) error {
	i.Status = ImportStatusFinished
	i.Yield = yield
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

// 获取实时设备的导入记录
// 生成以当前时间日期为结尾的key，通过key缓存获取数据
// 当缓存中没有该日期的实时导入记录时，从数据库获取
// 当数据库中没有该日期的实时导入记录时，创建新纪录
func (i *ImportRecord) DeviceRealtimeRecord(device *Device) error {
	// 生成key
	cacheKey := i.genKey(device.ID)

	// 获取缓存
	cacheValue := cache.Get(cacheKey)
	if cacheValue != nil {
		record, ok := cacheValue.(ImportRecord)
		if ok {
			if err := copier.Copy(i, &record); err == nil {
				return nil
			}
		}
	}

	var query = "device_id = ? AND import_type = ? AND DATE(created_at) = DATE(?) AND import_records.status = ?"

	// 查询数据库 [当前设备的 实时导入的 当前日期的 正在导入的] 导入记录
	var record ImportRecord
	if err := Model(&ImportRecord{}).Where(
		query, device.ID, ImportRecordTypeRealtime, time.Now(), ImportStatusImporting,
	).Find(&record).Error; err == nil {
		if err := copier.Copy(i, &record); err == nil {
			return nil
		}
	}

	// 更新当前版本的总数及良率统计
	{
		var lastRealtimeRecord ImportRecord
		if err := Model(&ImportRecord{}).Where(
			query, device.ID, ImportRecordTypeRealtime, time.Now().AddDate(0, 0, -1), ImportStatusImporting,
		).Find(&lastRealtimeRecord).Error; err == nil {
			var version MaterialVersion
			if err := version.Get(lastRealtimeRecord.MaterialVersionID); err == nil {
				_ = version.UpdateWithRecord(&lastRealtimeRecord)
			}
			lastRealtimeRecord.Status = ImportStatusFinished
			_ = Save(&lastRealtimeRecord)
		}
	}

	// 创建新记录
	i.MaterialID = device.MaterialID
	i.Status = ImportStatusImporting
	i.DeviceID = device.ID
	i.Path = "realtime"
	i.ImportType = ImportRecordTypeRealtime
	if err := Create(i).Error; err != nil {
		return err
	}

	_ = cache.Set(cacheKey, *i)
	return nil
}
