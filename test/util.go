package test

import (
	"github.com/SasukeBo/ftpviewer/conf"
	"github.com/jinzhu/gorm"

	// db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// ClearDB clear database
func ClearDB() {
	db, _ := gorm.Open("mysql", conf.DBdnstest)

	db.Exec("delete from devices where 1")
	db.Exec("delete from materials where 1")
	db.Exec("delete from products where 1")
	db.Exec("delete from size_values where 1")
	db.Exec("delete from sizes where 1")
	db.Exec("delete from system_configs where 1")
	db.Exec("delete from users where 1")
}

// SetConfig _
func SetConfig() {
	db, _ := gorm.Open("mysql", conf.DBdnstest)

	db.Exec(`
	INSERT INTO system_configs (system_configs.key, system_configs.value, id)
	values
	("ftp_host", "localhost", 1),
	("ftp_port", "44762", 2),
	("ftp_username", "admin", 3),
	("ftp_password", "123456", 4)
	`)
}
