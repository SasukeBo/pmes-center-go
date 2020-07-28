package main

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	dbUser = "root"
	dbPass = "Wb922149@...S"
	dbHost = "192.168.5.146"
	dbPort = "44766"
	dbName = "pmes_data_center"

	materialID = 20
	versionID  = 6
)

type DecodeTemplate struct {
	gorm.Model
	MaterialID           uint `gorm:"not null"`
	MaterialVersionID    uint `gorm:"not null"` // 料号版本ID
	UserID               uint
	DataRowIndex         int
	CreatedAtColumnIndex int       `gorm:"not null"`
	ProductColumns       types.Map `gorm:"type:JSON;not null"`
	PointColumns         types.Map `gorm:"type:JSON;not null"`
}

type Point struct {
	ID                uint   `gorm:"primary_key;column:id"`
	Name              string `gorm:"not null"`
	MaterialID        uint   `gorm:"not null"`
	MaterialVersionID uint   `gorm:"not null"`
	Index             int    `gorm:"not null"`
	UpperLimit        float64
	LowerLimit        float64
	Nominal           float64
}

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

func main() {
	uri := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPass,
		dbHost,
		dbPort,
		dbName,
	)
	conn, err := gorm.Open("mysql", uri)
	if err != nil {
		panic(err)
	}
	conn.LogMode(true)

	tx := conn.Begin()

	//var version MaterialVersion
	//err = conn.Model(&MaterialVersion{}).Where("material_id = ? AND active = true", materialID).Find(&version).Error
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = tx.Exec("update points set material_version_id = ? where material_id = ?",
	//	version.ID, materialID,
	//).Error
	//if err != nil {
	//	tx.Rollback()
	//	panic(err)
	//}
	//
	//err = tx.Exec(
	//	"update decode_templates set material_version_id = ? where material_id = ? AND decode_templates.default = true",
	//	version.ID, materialID,
	//).Error
	//if err != nil {
	//	tx.Rollback()
	//	panic(err)
	//}
	//
	//err = tx.Exec(
	//	"update import_records set material_version_id = ? where material_id = ?",
	//	version.ID, materialID,
	//).Error
	//if err != nil {
	//	tx.Rollback()
	//	panic(err)
	//}
	//
	//err = tx.Exec(
	//	"update products set material_version_id = ? where material_id = ?",
	//	version.ID, materialID,
	//).Error
	//if err != nil {
	//	tx.Rollback()
	//	panic(err)
	//}

	var template DecodeTemplate
	err = tx.Model(&DecodeTemplate{}).Where("material_id = ? AND material_version_id = ?", materialID, versionID).Find(&template).Error
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	var points []Point
	err = tx.Model(&Point{}).Where("material_id = ? AND material_version_id = ?", materialID, versionID).Find(&points).Error
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	for _, p := range points {
		i, ok := template.PointColumns[p.Name]
		if !ok {
			continue
		}
		index, ok := i.(float64)
		if !ok {
			fmt.Printf("index %v is not float64\n", i)
			continue
		}

		p.Index = int(index)
		err := tx.Save(&p).Error
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	tx.Commit()
}
