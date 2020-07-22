package orm

import (
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/jinzhu/gorm"
)

// MaterialVersion 材料版本号
type MaterialVersion struct {
	gorm.Model
	Version     string `gorm:"not null"`
	Description string
	MaterialID  uint `gorm:"not null"`
	Active      bool `gorm:"default:false"`
	UserID      uint
}

func (mv *MaterialVersion) Get(id uint) *errormap.Error {
	if err := DB.Model(mv).Where("id = ?", id).First(mv).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}
