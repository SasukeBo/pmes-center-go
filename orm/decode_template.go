package orm

import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/cache"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"strings"
)

// 数据文件解析模板
// 用于指定文件的必要数据位置

type DecodeTemplate struct {
	gorm.Model
	Name                 string `gorm:"not null"`
	MaterialID           uint   `gorm:"not null"`
	UserID               uint
	Description          string
	DataRowIndex         int
	CreatedAtColumnIndex int       `gorm:"not null"`
	ProductColumns       types.Map `gorm:"type:JSON;not null"`
	PointColumns         types.Map `gorm:"type:JSON;not null"`
	Default              bool      `gorm:"default:false"` // 标识是否为默认模板
}

const decodeTemplateCacheKey = "cache_decode_template_%v_%v"

/*	callbacks
--------------------------------------------------------------------------------------------------------------------- */

// 清除缓存
func (d *DecodeTemplate) AfterUpdate() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}
func (d *DecodeTemplate) AfterDelete() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}
func (d *DecodeTemplate) AfterSave() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID))
	return nil
}

// 清除缓存

/*	functions
--------------------------------------------------------------------------------------------------------------------- */
func (d *DecodeTemplate) GenDefaultProductColumns() (int, error) {
	productColumns := make(types.Map)
	var config SystemConfig
	err := config.GetConfig(SystemConfigProductColumnHeadersKey)
	if err != nil {
		return 0, err
	}

	headers := strings.Split(config.Value, ";")
	var columns []Column
	for i, header := range headers {
		vs := strings.Split(header, ":")
		columns = append(columns, Column{
			Name:  vs[0],
			Type:  vs[1],
			Index: i,
		})
	}

	productColumns["columns"] = columns
	d.ProductColumns = productColumns
	return len(columns), nil
}

func (d *DecodeTemplate) Get(id uint) *errormap.Error {
	if err := Model(d).Where("id = ?", id).First(d).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (d *DecodeTemplate) GetCache(id uint) *errormap.Error {
	var cacheKey = fmt.Sprintf(decodeTemplateCacheKey, "id", d.ID)
	cacheValue := cache.Get(cacheKey)
	if cacheValue != nil {
		template, ok := cacheValue.(DecodeTemplate)
		if ok {
			if err := copier.Copy(d, &template); err == nil {
				return nil
			}
		}
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

type Column struct {
	Name  string
	Index int
	Type  string
}

type ColumnValue struct {
	Label string
	Value interface{}
}
