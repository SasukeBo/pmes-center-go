package orm

import "time"

// File 存储已加载数据的文件路径
type File struct {
	ID           int
	Path         string
	MaterialID   int
	TotalRows    int
	FinishedRows int
	Finished     bool `gorm:"default:false"`
	FileDate     time.Time
}
