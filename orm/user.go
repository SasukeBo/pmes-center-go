package orm

import "github.com/jinzhu/gorm"

// User 系统用户
type User struct {
	gorm.Model
	Admin       bool   `gorm:"default:false"`
	Username    string `gorm:"not null;unique_index"`
	Password    string `gorm:"not null"`
	AccessToken string
}
