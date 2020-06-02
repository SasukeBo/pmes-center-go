package orm

import "github.com/jinzhu/gorm"

// 检测点位
// 产品的检测点位

// Point 点位
type Point struct {
	gorm.Model
	Name        string `gorm:"unique_index:uidx_name_material_id;not null"`
	MaterialID  uint   `gorm:"unique_index:uidx_name_material_id;not null"`
	UpperLimit  float64
	LowerLimit  float64
	Nominal     float64
}

// NotValid 校验数据有效性
func (p *Point) NotValid(v float64) bool {
	return p.Nominal > 0 && v > p.Nominal*100
}
