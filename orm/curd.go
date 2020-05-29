package orm

import "github.com/jinzhu/gorm"

func Create(object interface{}) *gorm.DB {
	return DB.Create(object)
}
