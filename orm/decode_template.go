package orm

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/gorm"
)

// 数据文件解析模板
// 用于指定文件的必要数据位置

type DecodeTemplate struct {
	gorm.Model
	MaterialID           uint `gorm:"not null"`
	MaterialVersionID    uint `gorm:"not null"` // 料号版本ID
	UserID               uint
	DataRowIndex         int
	DeviceNameRowIndex   int       `gorm:"default:2"` // 默认在第二行
	CreatedAtColumnIndex int       `gorm:"not null"`  // 检测时间位置
	BarCodeIndex         int       // 编码读取位置
	BarCodeRuleID        uint      `gorm:"COMMENT:'编码规则ID';column:bar_code_rule_id"`
	ProductColumns       types.Map `gorm:"type:JSON;not null"`
}

const decodeTemplateCacheKey = "cache_decode_template_%v_%v"

/*	callbacks
--------------------------------------------------------------------------------------------------------------------- */

// 清除缓存
func (d *DecodeTemplate) AfterUpdate() error {
	_ = cache.Del(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}
func (d *DecodeTemplate) AfterDelete() error {
	_ = cache.Del(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}
func (d *DecodeTemplate) AfterSave() error {
	_ = cache.Del(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}

// 清除缓存

/*	functions
--------------------------------------------------------------------------------------------------------------------- */
func (d *DecodeTemplate) Get(id uint) *errormap.Error {
	if err := Model(d).Where("id = ?", id).First(d).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (d *DecodeTemplate) GetByVersionID(versionID uint) *errormap.Error {
	if err := Model(d).Where("material_version_id = ?", versionID).First(d).Error; err != nil {
		return handleError(err, "material_version_id", versionID)
	}

	return nil
}

func (d *DecodeTemplate) GetCache(id uint) *errormap.Error {
	var cacheKey = fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID)
	err := cache.Scan(cacheKey, d)
	if err == nil {
		return nil
	}

	if err := DB.Model(d).Where("id = ?", id).First(d).Error; err != nil {
		return handleError(err, "id", id)
	}
	_ = cache.Set(cacheKey, *d)
	return nil
}

const (
	ProductColumnTypeString   = "String"
	ProductColumnTypeInteger  = "Integer"
	ProductColumnTypeFloat    = "Float"
	ProductColumnTypeDatetime = "Datetime"
)

// Column template product column struct
type Column struct {
	Prefix string
	Token  string
	Label  string
	Index  int
	Type   string
}
