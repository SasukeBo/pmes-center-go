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
	db.Exec("delete from point_values where 1")
	db.Exec("delete from points where 1")
	db.Exec("delete from sizes where 1")
	db.Exec("delete from system_configs where 1")
	db.Exec("delete from users where 1")
}

// SetConfig _
//func SetConfig() {
//	orm.DB.Exec(`
//	INSERT INTO system_configs (system_configs.key, system_configs.value, id)
//	values
//	("ftp_host", "localhost", 1),
//	("ftp_port", "44762", 2),
//	("ftp_username", "admin", 3),
//	("ftp_password", "123456", 4)
//	`)
//}
//
//// SetMaterialAndDevice _
//func SetMaterial(name string, deviceID int) {
//	orm.DB.Exec(`INSERT INTO materials (materials.name) values (?)`, name)
//	orm.DB.Exec(`INSERT INTO devices (material_id, name) values (?, ?)`, name, fmt.Sprintf("%s设备%d", name, deviceID))
//}
