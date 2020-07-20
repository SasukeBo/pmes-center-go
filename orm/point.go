package orm

import (
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"os"
	"path/filepath"
)

// 检测点位
// 产品的检测点位

func init() {
	DB.Model(&Point{}).AddUniqueIndex("unique_idx_point_name_material_id_version", "material_id", "material_version_id", "name")
}

// Point 点位
type Point struct {
	ID                uint   `gorm:"primary_key;column:id"`
	Name              string `gorm:"not null"`
	MaterialID        uint   `gorm:"not null"`
	MaterialVersionID uint   `gorm:"not null"`
	UpperLimit        float64
	LowerLimit        float64
	Nominal           float64
}

// NotValid 校验数据有效性
func (p *Point) NotValid(v float64) bool {
	return p.Nominal > 0 && v > p.Nominal*100
}

func (p *Point) Get(id uint) *errormap.Error {
	if err := DB.Model(p).Where("id = ?", id).First(p).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func setupPointsImportTemplate() {
	var iptFile File
	var token = configer.GetString("points_import_template_token")
	if err := iptFile.GetByToken(token); err == nil {
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	filePath := filepath.Join(pwd, "priv/import_points_template.xlsx")
	iptFile = File{
		Token:       token,
		Name:        "检测点位导入模板.xlsx",
		Path:        filePath,
		ContentType: XlsxContentType,
	}

	if err := Create(&iptFile).Error; err != nil {
		panic(err)
	}
}
