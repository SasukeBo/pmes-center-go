package orm

import (
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/jinzhu/gorm"
)

// Material 材料
type Material struct {
	gorm.Model
	Name          string  `gorm:"not null"`
	YieldScore    float64 // 良率百分比目标线
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

func (m *Material) GetCurrentVersion() (*MaterialVersion, error) {
	var version MaterialVersion
	if err := Model(&MaterialVersion{}).Where("material_id = ? AND material_versions.active = true", m.ID).First(&version).Error; err != nil {
		return nil, err
	}

	return &version, nil
}

func (m *Material) GetCurrentTemplateDecodeRule() *BarCodeRule {
	var template DecodeTemplate
	query := Model(&DecodeTemplate{}).Joins("JOIN material_versions ON decode_templates.material_version_id = material_versions.id")
	query = query.Where("decode_templates.material_id = ? AND material_versions.active = true", m.ID)
	if err := query.Find(&template).Error; err != nil {
		log.Errorln(err)
		return nil
	}

	var rule BarCodeRule
	if err := rule.Get(template.BarCodeRuleID); err != nil {
		log.Errorln(err)
		return nil
	}

	return &rule
}

func (m *Material) GetCurrentVersionTemplate() (*DecodeTemplate, error) {
	version, err := m.GetCurrentVersion()
	if err != nil {
		return nil, err
	}

	var template DecodeTemplate
	if err := template.GetByVersionID(version.ID); err != nil {
		return nil, err
	}

	return &template, nil
}
