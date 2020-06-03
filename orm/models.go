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

const (
	SystemConfigFtpHostKey              = "ftp_host"
	SystemConfigFtpPortKey              = "ftp_port"
	SystemConfigFtpUsernameKey          = "ftp_username"
	SystemConfigFtpPasswordKey          = "ftp_password"
	SystemConfigProductColumnHeadersKey = "product_column_headers"
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

func GenerateDefaultConfig() {
	SetIfNotExist(SystemConfigFtpHostKey)
	SetIfNotExist(SystemConfigFtpPortKey)
	SetIfNotExist(SystemConfigFtpUsernameKey)
	SetIfNotExist(SystemConfigFtpPasswordKey)
	SetIfNotExist(SystemConfigProductColumnHeadersKey)
}

func alterTableUtf8(tbName string) {
	Exec(fmt.Sprintf("ALTER TABLE %s CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci", tbName))
}

func utf8GeneralCI(tableNames []string) {
	Exec("SET collation_connection = 'utf8_general_ci'")
	Exec(fmt.Sprintf("ALTER DATABASE %s CHARACTER SET utf8 COLLATE utf8_general_ci", configer.GetString("db_name")))
	for _, name := range tableNames {
		alterTableUtf8(name)
	}
}

func generateRootUser() {
	account := configer.GetString("root_name")
	var root User
	err := Model(&User{}).Where("account = ?", account).First(&root).Error
	if err != nil {
		root = User{
			IsAdmin:  true,
			Account:  account,
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
	DB.LogMode(false)
	env := configer.GetString("env")

	err = DB.AutoMigrate(
		&DecodeTemplate{},
		&Device{},
		&ImportRecord{},
		&Material{},
		&Point{},
		&Product{},
		&SystemConfig{},
		&User{},
		&UserLogin{},
		&UserRole{},
	).Error
	if err != nil {
		panic(fmt.Errorf("migrate to db error: \n%v", err.Error()))
	}

	if env != "test" || env != "TEST" {
		generateRootUser()
		GenerateDefaultConfig()
		tableNames := []string{"decode_templates", "devices", "import_records", "materials", "points", "products", "system_configs", "users"}
		utf8GeneralCI(tableNames)
	}

	if env == "prod" {
		DB.LogMode(false)
	} else {
		DB.LogMode(true)
	}
}
