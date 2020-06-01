package orm

import "github.com/jinzhu/gorm"

// Material 材料
type Material struct {
	gorm.Model
	Name          string `gorm:"not null;unique_index"`
	CustomerCode  string
	ProjectRemark string
}

func (m *Material) GetWithName(name string) error {
	return DB.Model(m).Where("name = ?", name).First(m).Error
}

func (m *Material) GetDefaultTemplate() (*DecodeTemplate, error) {
	var template DecodeTemplate
	err := DB.Model(&template).Where("material_id = ? AND default = 1", m.ID).First(&template).Error
	return &template, err
}
