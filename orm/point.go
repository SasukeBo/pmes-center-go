package orm

// 检测点位
// 产品的检测点位

// Point 点位
type Point struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	MaterialID int    `gorm:"not null;index"`
	UpperLimit float64
	LowerLimit float64
	Nominal    float64
}

// NotValid 校验数据有效性
func (p *Point) NotValid(v float64) bool {
	return p.Nominal > 0 && v > p.Nominal*100
}
