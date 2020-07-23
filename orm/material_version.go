package orm

import (
	"errors"
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
	Amount      int
	Yield       float64
}

func (mv *MaterialVersion) Get(id uint) *errormap.Error {
	if err := DB.Model(mv).Where("id = ?", id).First(mv).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (mv *MaterialVersion) UpdateWithRecord(record *ImportRecord) error {
	if mv == nil {
		return errors.New("cannot update <nil> version")
	}
	currentTotal := mv.Amount
	currentTotal = currentTotal + record.RowFinishedCount
	mv.Amount = currentTotal

	importOK := int(float64(record.RowFinishedCount) * record.Yield)
	currentOK := int(float64(currentTotal) * mv.Yield)
	currentOK = currentOK + importOK

	if currentTotal == 0 {
		mv.Yield = 0
	} else {
		mv.Yield = float64(currentOK) / float64(currentTotal)
	}
	return Save(mv).Error
}
