package main

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"time"
)

/*
该文件用于拆分product表为三个表，将JSON数据类型单独存表
目的为了加快产品数据的查询速度
*/

var (
	dbUser = "root"
	dbPass = "Wb922149@...S"
	dbHost = "sasuke.local"
	dbPort = "3306"
	dbName = "pmes_data_center"
)

type NewProduct struct {
	ID                uint      `gorm:"column:id;primary_key"`
	ImportRecordID    uint      `gorm:"COMMENT:'导入记录ID';column:import_record_id;not null;index"`
	MaterialVersionID uint      `gorm:"COMMENT:'料号版本ID';index"`
	MaterialID        uint      `gorm:"COMMENT:'料号ID';column:material_id;not null;index"`
	DeviceID          uint      `gorm:"COMMENT:'检测设备ID';column:device_id;not null;index"`
	Qualified         bool      `gorm:"COMMENT:'产品尺寸是否合格';column:qualified;default:false"`
	BarCode           string    `gorm:"COMMENT:'识别条码';column:bar_code;"`
	BarCodeStatus     int       `gorm:"COMMENT:'条码解析状态';column:bar_code_status;default:1"`
	CreatedAt         time.Time `gorm:"COMMENT:'产品检测时间';index"` // 检测时间
}

type ProductPoint struct {
	ID        uint      `gorm:"column:id;primary_key"`
	ProductID uint      `gorm:"column:product_id;index"`
	Points    types.Map `gorm:"COMMENT:'产品点位检测值集合';type:JSON;not null"`
}

type ProductAttribute struct {
	ID         uint      `gorm:"column:id;primary_key"`
	ProductID  uint      `gorm:"column:product_id;index"`
	Attributes types.Map `gorm:"COMMENT:'产品属性值集合';type:JSON;not null"`
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
	conn.LogMode(false)
	orm.DB = conn

	err = orm.DB.AutoMigrate(
		&NewProduct{},
		&ProductPoint{},
		&ProductAttribute{},
	).Error

	var products []orm.Product
	if err := orm.Model(&orm.Product{}).Where("id > ?", 3169179).Find(&products).Error; err == nil {
		for i, p := range products {
			var np NewProduct
			copier.Copy(&np, &p)
			np.ID = 0
			orm.Save(&np)
			orm.Save(&ProductPoint{
				ProductID: np.ID,
				Points:    p.PointValues,
			})
			orm.Save(&ProductAttribute{
				ProductID:  np.ID,
				Attributes: p.Attribute,
			})
			fmt.Printf("finish %d, product id %v\n", i, p.ID)
		}
	}
}
