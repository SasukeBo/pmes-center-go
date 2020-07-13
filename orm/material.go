package orm

import (
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/jinzhu/gorm"
)

// Material 材料
type Material struct {
	gorm.Model
	Name          string `gorm:"not null"`
	CustomerCode  string
	ProjectRemark string
}

func (m *Material) GetWithName(name string) *errormap.Error {
	if err := DB.Model(m).Where("name = ?", name).First(m).Error; err != nil {
		return handleError(err, "name", name)
	}

	return nil
}

func (m *Material) Get(id uint) *errormap.Error {
	if err := DB.Model(m).Where("id = ?", id).First(m).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (m *Material) GetDefaultTemplate() (*DecodeTemplate, error) {
	var template DecodeTemplate
	err := DB.Model(&template).Where("material_id = ? AND `decode_templates`.`default` = ?", m.ID, true).First(&template).Error
	return &template, err
}
