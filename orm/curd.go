package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

// GetSystemConfig 获取缓存
func GetSystemConfig(key string) *SystemConfig {
	var conf SystemConfig
	if err := DB.Where("system_configs.key = ?", key).Find(&conf).Error; err != nil {
		return nil
	}
	return &conf
}

// GetUserWithToken 获取缓存
func GetUserWithToken(token string) *User {
	if token == "" {
		return nil
	}
	var user User
	if err := DB.Where("access_token = ?", token).Find(&user).Error; err != nil {
		return nil
	}
	return &user
}

// GetMaterialWithID _
func GetMaterialWithID(materialID int) *Material {
	var m Material
	if err := DB.Where("id = ?", materialID).First(&m).Error; err != nil {
		return nil
	}
	return &m
}

// GetMaterialWithName _
func GetMaterialWithName(name string) *Material {
	var m Material
	if err := DB.Where("name = ?", name).First(&m).Error; err != nil {
		return nil
	}
	return &m
}

// GetDeviceWithName _
func GetDeviceWithName(name string) *Device {
	var d Device
	if err := DB.Where("name = ?", name).First(&d).Error; err != nil {
		return nil
	}
	return &d
}

// GetDeviceWithID _
func GetDeviceWithID(id int) *Device {
	var d Device
	if err := DB.Where("id = ?", id).First(&d).Error; err != nil {
		return nil
	}
	return &d
}

// GetSizeWithMaterialIDSizeName _
func GetSizeWithMaterialIDSizeName(sn string, materialID int, tx *gorm.DB) *Size {
	conn := DB
	if tx != nil {
		conn = tx
	}

	var s Size
	if err := conn.Where("name = ? AND material_id = ?", sn, materialID).First(&s).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			fmt.Println(err)
		}
		return nil
	}
	return &s
}

// GetPointWithSizeIDPointName _
func GetPointWithSizeIDPointName(pn string, sizeID int, tx *gorm.DB) *Point {
	conn := DB
	if tx != nil {
		conn = tx
	}

	var p Point
	if err := conn.Where("name = ? AND size_id = ?", pn, sizeID).First(&p).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			fmt.Println(err)
		}
		return nil
	}
	return &p
}

func GetSizeWithID(id int) *Size {
	var s Size
	if err := DB.Where("id = ?", id).Find(&s).Error; err != nil {
		return nil
	}
	return &s
}

func GetFileListWithPath(path string) *File {
	var fl File
	if err := DB.Where("path = ?", path).First(&fl).Error; err != nil {
		return nil
	}
	return &fl
}
