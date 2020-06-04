package orm

import (
	"github.com/SasukeBo/ftpviewer/orm/types"
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

func (d *DecodeTemplate) GenDefaultProductColumns() error {
	productColumns := make(types.Map)
	var config SystemConfig
	err := config.GetConfig(SystemConfigProductColumnHeadersKey)
	if err != nil {
		return err
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
	return nil
}

func (d *DecodeTemplate) Get(id uint) error {
	return Model(d).Where("id = ?", id).First(d).Error
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
