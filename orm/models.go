package models

import (
	"github.com/jinzhu/gorm"
)

type SystemConfig struct {
	gorm.Model
	Key   string `gorm:""`
	Value string `gorm:""`
}
