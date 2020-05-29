package orm

// Point 点位
type Point struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	SizeID     int    `gorm:"column:size_id;not null;index"`
	Index      int    `gorm:"not null"`
	UpperLimit float64
	LowerLimit float64
	Nominal    float64
}

// NotValid 校验数据有效性
func (p *Point) NotValid(v float64) bool {
	return p.Nominal > 0 && v > p.Nominal*100
}

// PointValue 点位值
type PointValue struct {
	PointID     int     `gorm:"column:point_id;not null;index"`
	ProductUUID string  `gorm:"column:product_uuid;not null;index"`
	V           float64 `gorm:"column:v;not null"`
}
