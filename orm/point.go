package orm

import (
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// 检测点位
// 产品的检测点位

// Point 点位
type Point struct {
	ID         uint   `gorm:"primary_key;column:id"`
	Name       string `gorm:"unique_index:uidx_name_material_id;not null"`
	MaterialID uint   `gorm:"unique_index:uidx_name_material_id;not null"`
	UpperLimit float64
	LowerLimit float64
	Nominal    float64
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
	iptFile.GetByToken(token)

	var fileCachePath = configer.GetString("file_cache_path")
	p := path.Join(fileCachePath, "priv", "templates")
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		panic("cannot create templates directory.")
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tplFile, err := os.Open(filepath.Join(pwd, "priv/import_points_template.xlsx"))
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(tplFile)
	if err != nil {
		panic(err)
	}

	size := len(content)

	err = ioutil.WriteFile(path.Join(p, "import_points_template.xlsx"), content, 0644)
	if err != nil {
		panic(err)
	}

	iptFile.Token = token
	iptFile.Name = "检测点位导入模板.xlsx"
	iptFile.Path = "/priv/templates/import_points_template.xlsx"
	iptFile.Size = uint(size)
	iptFile.ContentType = xlsxContentType

	if err := Save(&iptFile).Error; err != nil {
		panic(err)
	}
}
