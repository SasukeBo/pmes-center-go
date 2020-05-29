package orm

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/gorm"
	"time"

	// set db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB connection to database
var DB *gorm.DB

func createUriWithDBName(name string) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		configer.GetString("db_user"),
		configer.GetString("db_pass"),
		configer.GetString("db_host"),
		configer.GetString("db_port"),
		name,
	)
}

func generateDefaultConfig() {
	SetIfNotExist("ftp_host", configer.GetString("ftp_host"))
	SetIfNotExist("ftp_port", configer.GetString("ftp_port"))
	SetIfNotExist("ftp_username", configer.GetString("ftp_username"))
	SetIfNotExist("ftp_password", configer.GetString("ftp_password"))
}

func alterTableUtf8(tbname string) {
	DB.Exec(fmt.Sprintf("ALTER TABLE %s CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci", tbname))
}

func utf8GeneralCI(tableNames []string) {
	DB.Exec("SET collation_connection = 'utf8_general_ci'")
	DB.Exec(fmt.Sprintf("ALTER DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci", configer.GetString("db_name")))
	for _, name := range tableNames {
		alterTableUtf8(name)
	}
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

func init() {
	var err error
	var uri = createUriWithDBName("mysql")
	var dbname = configer.GetString("db_name")

	reconnectLimit := 5
	for {
		conn, err := gorm.Open("mysql", uri)
		if err != nil && reconnectLimit > 0 {
			log.Errorln(err)
			reconnectLimit--
			time.Sleep(time.Duration(5-reconnectLimit) * 2 * time.Second)
			log.Info("open connection with %s failed, try again ...\n", uri)
			continue
		}
		conn.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
		conn.Close()
		break
	}

	DB, err = gorm.Open("mysql", createUriWithDBName(dbname))
	if err != nil {
		panic(err)
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
		&DecodeTemplate{},
		&Device{},
		&ImportRecord{},
		&Material{},
		&Point{},
		&Product{},
		&SystemConfig{},
		&User{},
	).Error
	if err != nil {
		panic(fmt.Errorf("migrate to db error: \n%v", err.Error()))
	}

	if configer.GetString("env") != "test" {
		generateRootUser()
	}
	generateDefaultConfig()

	tableNames := []string{"decode_templates", "devices", "import_records", "materials", "points", "products", "system_configs", "users"}
	utf8GeneralCI(tableNames)
}
