package main

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/gorm"
)

/*
该文件用于将旧数据中的barCode解析为产品数据的attribute
*/

var (
	dbUser = "root"
	dbPass = "Wb922149@...S"
	dbHost = "192.168.5.146"
	dbPort = "44766"
	dbName = "pmes_data_center"

	materialID = 2
	productID  = 514539
)

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

	orm.DB = conn
	tx := orm.Begin()

	var material orm.Material
	err = orm.Model(&orm.Material{}).Where("id = ?", materialID).Find(&material).Error
	if err != nil {
		panic(err)
	}
	rule := material.GetCurrentTemplateDecodeRule()
	var decoder *logic.BarCodeDecoder
	fmt.Printf("%+v\n", rule)
	if rule != nil {
		decoder = logic.NewBarCodeDecoder(rule)
	}
	if decoder == nil {
		panic("nil decoder")
	}

	var products []orm.Product
	query := orm.Model(&orm.Product{}).Where("material_id = ?", materialID)
	query = query.Where("id = ?", productID)
	err = query.Find(&products).Error
	if err != nil {
		panic(err)
	}

	for _, p := range products {
		attributes, status := decoder.Decode(p.BarCode)
		p.Attribute = attributes
		p.BarCodeStatus = status
		err := tx.Save(&p).Error
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	tx.Commit()
}
