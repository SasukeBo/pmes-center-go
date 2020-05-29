package orm

// Material 材料
type Material struct {
	ID            int    `gorm:"column:id;primary_key"`
	Name          string `gorm:"not null;unique_index"`
	CustomerCode  string
	ProjectRemark string
}
