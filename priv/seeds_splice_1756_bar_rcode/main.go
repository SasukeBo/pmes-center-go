package main

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

var (
	dbUser = "root"
	dbPass = "Wb922149@...S"
	dbHost = "192.168.5.146"
	dbPort = "44766"
	dbName = "pmes_data_center"

	materialID = 2
	versionID  = 6
)

// Product 产品表
type Product struct {
	ID                uint      `gorm:"column:id;primary_key"`
	ImportRecordID    uint      `gorm:"column:import_record_id;not null;index"`
	MaterialVersionID uint      `gorm:"index"`
	MaterialID        uint      `gorm:"column:material_id;not null;index"`
	DeviceID          uint      `gorm:"column:device_id;not null;index"`
	Qualified         bool      `gorm:"column:qualified;default:false"`
	BarCode           string    `gorm:"column:bar_code;"`
	CreatedAt         time.Time `gorm:"index"` // 检测时间
	Attribute         types.Map `gorm:"type:JSON;not null"`
	PointValues       types.Map `gorm:"type:JSON;not null"`
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

	var products []Product
	err = tx.Model(&Product{}).Where("material_id = ? AND bar_code is NULL", 2).Find(&products).Error
	if err != nil {
		tx.Rollback()
		panic(err)
	}

	for i, p := range products {
		p.BarCode = fmt.Sprint(p.Attribute["Line"])
		attr := make(types.Map)
		if len(p.BarCode) > 0 {
			attr["DayCode"] = p.BarCode[:1]
		}
		if len(p.BarCode) > 1 {
			attr["Line"] = p.BarCode[1:2]
		}
		p.Attribute = attr
		if err := tx.Save(&p).Error; err != nil {
			tx.Rollback()
			panic(err)
		}
		fmt.Println(i)
	}

	tx.Commit()
}
