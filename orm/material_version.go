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

func (mv *MaterialVersion) GetActiveWithMaterialID(id uint) *errormap.Error {
	if err := DB.Model(mv).Where("active = true AND material_id = ?", id).First(mv).Error; err != nil {
		return handleError(err, "material_id", id)
	}

	return nil
}

func (mv *MaterialVersion) GetTemplate() (*DecodeTemplate, error) {
	var template DecodeTemplate
	if err := Model(DecodeTemplate{}).Where("material_version_id = ?", mv.ID).Find(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

func (mv *MaterialVersion) UpdateWithRecord(record *ImportRecord, conn ...*gorm.DB) error {
	db := choseConn(conn...)

	if mv == nil {
		return errors.New("cannot update <nil> version")
	}

	switch record.Status {
	case ImportStatusFinished:
		currentTotal := mv.Amount
		total := currentTotal + record.RowFinishedCount
		mv.Amount = total

		importOK := int(float64(record.RowFinishedCount) * record.Yield)
		currentOK := int(float64(currentTotal) * mv.Yield)
		currentOK = currentOK + importOK

		if total == 0 {
			mv.Yield = 0
		} else {
			mv.Yield = float64(currentOK) / float64(total)
		}
		return db.Save(mv).Error

	case ImportStatusReverted:
		currentTotal := mv.Amount
		currentOK := int(float64(currentTotal) * mv.Yield)

		recordOK := int(float64(record.RowFinishedCount) * record.Yield)

		total := currentTotal - record.RowFinishedCount
		ok := currentOK - recordOK
		mv.Amount = total

		if total == 0 {
			mv.Yield = 0
		} else {
			mv.Yield = float64(ok) / float64(total)
		}
		return db.Save(mv).Error
	}

	return nil
}
