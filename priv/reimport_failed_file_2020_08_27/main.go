package main

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"path/filepath"
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
	var files []orm.File
	err = orm.Model(&orm.File{}).Where("id <= 1434 AND id >= 1338").Find(&files).Error
	if err != nil {
		panic(err)
	}

	var base = "/home/sasukebo/pmes_data_center_file_cache"
	for _, f := range files {
		content, err := ioutil.ReadFile(filepath.Join(base, f.Path))
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err := ioutil.WriteFile(filepath.Join("/home/sasukebo/pmesdata/lost_xlsx", f.Name), content, 0644); err != nil {
			fmt.Println(err)
		}
	}
}
