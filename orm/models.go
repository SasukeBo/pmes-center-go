package orm

import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/conf"
	"github.com/jinzhu/gorm"
	// set db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

// DB connection to database
var DB *gorm.DB

// SystemConfig 系统设置表
type SystemConfig struct {
	gorm.Model
	Key   string
	Value string
}

// Device 生产设备表
type Device struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"not null"`
	MaterialID string `gorm:"column:material_id;not null;index"`
}

// Product 产品表
type Product struct {
	ID         int    `gorm:"column:id;primary_key"`
	UUID       string `gorm:"column:product_uuid;unique_index;not null"`
	MaterialID string `gorm:"column:material_id;not null;index"`
	DeviceID   int    `gorm:"column:device_id;not null"`
	CreatedAt  time.Time
}

// Size 尺寸
type Size struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"unique_index;not null"`
	MaterialID string `gorm:"column:material_id;not null;index"`
	UpperLimit float64
	LowerLimit float64
}

// SizeValue 检测值
type SizeValue struct {
	ID          int
	SizeID      int `gorm:"column:size_id;not null"`
	ProductUUID int `gorm:"column:product_uuid;not null"`
	Value       float64
}

func init() {
	var err error
	DB, err = gorm.Open("mysql", conf.DBdns)
	if err != nil {
		panic(fmt.Errorf("open connection to db error: \n%v", err.Error()))
	}
	err = DB.AutoMigrate(
		&SystemConfig{},
		&Device{},
		&Product{},
		&Size{},
		&SizeValue{},
	).Error
	if err != nil {
		panic(fmt.Errorf("migrate to db error: \n%v", err.Error()))
	}
}
