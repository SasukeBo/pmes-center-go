package orm

import (
	"fmt"

	"github.com/SasukeBo/ftpviewer/conf"
	"github.com/jinzhu/gorm"

	"crypto/md5"
	"time"

	// set db driver
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// DB connection to database
var DB *gorm.DB

// User 系统用户
type User struct {
	gorm.Model
	Admin       bool   `gorm:"default:false"`
	Username    string `gorm:"not null;unique_index"`
	Password    string `gorm:"not null"`
	AccessToken string
}

// SystemConfig 系统设置表
type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"unique_index"`
	Value string
}

// Material 材料
type Material struct {
	ID   int    `gorm:"column:id;primary_key"`
	Name string `gorm:"not null;unique_index"`
}

// Device 生产设备表
type Device struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"not null;unique_index"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
}

// Product 产品表
type Product struct {
	ID         int    `gorm:"column:id;primary_key"`
	UUID       string `gorm:"column:product_uuid;unique_index;not null"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
	DeviceID   int    `gorm:"column:device_id;not null"`
	Qualified  bool   `gorm:"column:qualified;default:false"`
	CreatedAt  time.Time
}

// Size 尺寸
type Size struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	Index      int    `gorm:"column:index;not null"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
	UpperLimit float64
	LowerLimit float64
}

// SizeValue 检测值
type SizeValue struct {
	ID          int
	SizeID      int    `gorm:"column:size_id;index; not null"`
	DeviceID    int    `gorm:"column:device_id; not null"`
	ProductUUID string `gorm:"column:product_uuid;not null"`
	Value       float64
	Qualified   bool `gorm:"column:qualified;default:false"`
	CreatedAt   time.Time
}

// FileList 存储已加载数据的文件路径
type FileList struct {
	ID         int
	Path       string
	MaterialID int
	Finished   bool `gorm:"default:false"`
}

func init() {
	var err error

	if conf.GetEnv() == "TEST" {
		DB, err = gorm.Open("mysql", conf.DBdnstest)
	} else {
		DB, err = gorm.Open("mysql", conf.DBdns)
	}

	if err != nil {
		panic(fmt.Errorf("open connection to db error: \n%v", err.Error()))
	}
	DB.LogMode(true)
	err = DB.AutoMigrate(
		&SystemConfig{},
		&Device{},
		&Product{},
		&Size{},
		&SizeValue{},
		&Material{},
		&User{},
		&FileList{},
	).Error
	if err != nil {
		panic(fmt.Errorf("migrate to db error: \n%v", err.Error()))
	}

	generateRootUser()
	generateDefaultConfig()
}

func generateDefaultConfig() {
	if conf.GetEnv() == "TEST" {
		DB.Exec("DELETE FROM system_configs WHERE 1 = 1")
	}
	t := time.Now()
	var sql = `
	INSERT INTO system_configs (system_configs.key, system_configs.value, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	`
	DB.Exec(sql, "ftp_password", "123456", t, t)
	DB.Exec(sql, "ftp_username", "admin", t, t)
	DB.Exec(sql, "ftp_host", "0.0.0.0", t, t)
	DB.Exec(sql, "ftp_port", "44762", t, t)
	DB.Exec(sql, "cache_days", "30", t, t)
}

func generateRootUser() {
	var root User
	DB.Where("username = ?", "admin").First(&root)
	if root.ID > 0 {
		return
	}

	u := &User{
		Username: "admin",
		Password: Encrypt("admin"),
		Admin:    true,
	}

	if err := DB.Create(u).Error; err != nil {
		panic(err)
	}
}

// Encrypt _
func Encrypt(origin string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(origin)))
}
