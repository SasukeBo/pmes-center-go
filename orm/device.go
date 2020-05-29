package orm

// Device 生产设备表
type Device struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"not null;unique_index"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
}
