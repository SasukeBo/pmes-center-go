package orm

import "github.com/jinzhu/gorm"

func Create(object interface{}) *gorm.DB {
	return DB.Create(object)
}

func Save(object interface{}) *gorm.DB {
	return DB.Save(object)
}

func Delete(object interface{}, where ...interface{}) *gorm.DB {
	return DB.Delete(object, where...)
}

func Model(object interface{}) *gorm.DB {
	return DB.Model(object)
}

func Exec(sql string, variables ...interface{}) *gorm.DB {
	return DB.Exec(sql, variables...)
}

func Begin() *gorm.DB {
	return DB.Begin()
}
