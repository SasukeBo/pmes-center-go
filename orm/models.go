package orm

import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/SasukeBo/log"

	"github.com/SasukeBo/configer"
	"github.com/jinzhu/gorm"

	"time"

	// set db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB connection to database
var DB *gorm.DB

func Create(object interface{}) *gorm.DB {
	return DB.Create(object)
}

func init() {
	var err error
	var dns = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configer.GetString("db_user"),
		configer.GetString("db_pass"),
		configer.GetString("db_host"),
		configer.GetString("db_port"),
		configer.GetString("db_name"),
	)

	reconnectLimit := 5
	for {
		DB, err = gorm.Open("mysql", dns)
		log.Info("open connection to mysql %s\n", dns)
		if err != nil && reconnectLimit > 0 {
			log.Errorln(err)
			reconnectLimit--
			time.Sleep(time.Duration(5-reconnectLimit) * 2 * time.Second)
			log.Infoln("try to reconnect db again ...")
			continue
		}
		break
	}

	if configer.GetString("env") == "prod" {
		DB.LogMode(false)
	} else {
		DB.LogMode(true)
	}

	if err != nil {
		panic(fmt.Errorf("open connection to db error: \n%v", err.Error()))
	}

	err = DB.AutoMigrate(
		&SystemConfig{},
		&Device{},
		&Product{},
		&Size{},
		&Point{},
		&PointValue{},
		&Material{},
		&User{},
		&File{},
	).Error
	if err != nil {
		panic(fmt.Errorf("migrate to db error: \n%v", err.Error()))
	}

	if configer.GetString("env") != "test" {
		generateRootUser()
	}
	generateDefaultConfig()
	utf8GeneralCI()
}

func generateDefaultConfig() {
	SetIfNotExist("ftp_host", configer.GetString("ftp_host"))
	SetIfNotExist("ftp_port", configer.GetString("ftp_port"))
	SetIfNotExist("ftp_username", configer.GetString("ftp_username"))
	SetIfNotExist("ftp_password", configer.GetString("ftp_password"))
}

func utf8GeneralCI() {
	DB.Exec("SET collation_connection = 'utf8_general_ci'")
	DB.Exec("ALTER DATABASE ftpviewer CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE devices CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE files CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE materials CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE products CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE points CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE point_values CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE sizes CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE system_configs CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
	DB.Exec("ALTER TABLE users CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci")
}

func generateRootUser() {
	username := configer.GetString("root_name")
	var root User
	err := DB.Model(&User{}).Where("username = ?", username).First(&root).Error
	if err != nil {
		root = User{
			Admin:    true,
			Username: username,
			Password: util.Encrypt(configer.GetString("root_pass")),
		}
		err := DB.Create(&root).Error
		if err != nil {
			panic(fmt.Sprintf("Generate root user failed: %v", err))
		}
	}
}
