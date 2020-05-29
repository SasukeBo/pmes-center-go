package orm

// Size 尺寸
type Size struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
}

