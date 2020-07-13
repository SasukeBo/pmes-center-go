package orm

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// User 系统用户
type User struct {
	gorm.Model
	UUID     string `gorm:"unique_index;not null"`
	IsAdmin  bool   `gorm:"default:false"`
	Account  string `gorm:"not null;unique_index"`
	Password string `gorm:"not null"`
}

func (u *User) BeforeCreate() error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	u.UUID = uid.String()
	return nil
}

func (u *User) GetWithToken(token string) error {
	return DB.Model(u).Where("token = ?", token).First(u).Error
}

func (u *User) Get(id uint) error {
	return Model(u).Where("id = ?", id).First(u).Error
}

type UserLogin struct {
	gorm.Model
	UserID            uint `gorm:"not null"`
	AccessToken       string
	EncryptedPassword string `gorm:"column:encrypted_password"`
	IP                string
	KeepLogin         bool `gorm:"default: false"`
	UserAgent         string
}

type UserRole struct {
	UserID uint
	RoleID uint
}
