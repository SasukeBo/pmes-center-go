package orm

import (
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/gorm"
)

type BarCodeRule struct {
	gorm.Model
	CodeLength int       `gorm:"COMMENT:'编码长度';not null"`
	Name       string    `gorm:"COMMENT:'编码规则名称';not null;unique_index"` // 编码规则名称
	Remark     string    `gorm:"COMMENT:'编码规则描述';not null"`              // 规则描述
	UserID     uint      `gorm:"COMMENT:'编码规则创建人'"`                      // 创建人ID
	Items      types.Map `gorm:"COMMENT:'解析项配置';type:JSON;not null"`     // 存储解析规则
}

// BarCodeItem 二维码识别规则对象
// - IndexRange 表示识别码索引范围，为整型数组，大于等于两位时，取前两位索引范围的字符，一位时，取该位索引为字符，0位时忽略该规则。
// - Type 值类型，主要分两类Category和Date，前者一律处理为字符串，后者解析为日期
// - 当Type=Date时，有以下字段
// - DayCode - 日期编码，为字符串数组，长度应该大于等于2，前两位表示编码起始字符，按照1-9 A-Z的顺序，从第三个元素开始为剔除字符，即从编码起始
//   字符中剔除这些字符。当位数小于2时，视为无DayCode，则对应日期按照检测时间的日期补全。[1, Y, B, I, O] 表示从1到Y，去除B，I，O。
// - MonthCode - 月份编码，字符串数组，规则同DayCode。 [1, D, A] 表示从1到D，去除A。
type BarCodeItem struct {
	Label      string   `json:"label"`       // 解析项的名称，例如：冲压日期
	Key        string   `json:"key"`         // 解析项的英文标识，例如：ProduceDate
	IndexRange []int    `json:"index_range"` // 解析码索引区间，例如：[21,22]
	Type       string   `json:"type"`        // 解析项类型，例如：Datetime
	DayCode    []string `json:"day_code"`    // 日码区间
	MonthCode  []string `json:"month_code"`  // 月码区间
}

func (r *BarCodeRule) Get(id uint) *errormap.Error {
	if err := DB.Model(r).Where("id = ?", id).First(r).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

const (
	BarCodeStatusSuccess = 1 + iota
	BarCodeStatusIllegal
	BarCodeStatusEmpty
	BarCodeStatusTooShort
)
