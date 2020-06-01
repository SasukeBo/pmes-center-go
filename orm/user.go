package orm

import "github.com/jinzhu/gorm"

// User 系统用户
type User struct {
	gorm.Model
	IsAdmin       bool   `gorm:"default:false"`
	Username    string `gorm:"not null;unique_index"`
	Password    string `gorm:"not null"`
	AccessToken string
}

func (u *User) GetWithToken(token string) error {
	return DB.Model(u).Where("token = ?", token).First(u).Error
}

type UserRole struct {
	UserID uint
	RoleID uint
}
