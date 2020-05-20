package orm

import (
	"fmt"
	"os"

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
	ID            int    `gorm:"column:id;primary_key"`
	Name          string `gorm:"not null;unique_index"`
	CustomerCode  string
	ProjectRemark string
}

// Device 生产设备表
type Device struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"not null;unique_index"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
}

// Product 产品表
type Product struct {
	ID          int       `gorm:"column:id;primary_key"`
	UUID        string    `gorm:"column:uuid;unique_index;not null"`
	MaterialID  int       `gorm:"column:material_id;not null;index"`
	DeviceID    int       `gorm:"column:device_id;not null;index"`
	Qualified   bool      `gorm:"column:qualified;default:false"`
	CreatedAt   time.Time `gorm:"index"`
	D2Code      string    `gorm:"column:d2_code"`
	LineID      int       `gorm:"column:line_id;index"`
	JigID       string    `gorm:"column:jig_id;index"`
	MouldID     int       `gorm:"column:mould_id;index"`
	ShiftNumber int       `gorm:"index"`
}

// Size 尺寸
type Size struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	MaterialID int    `gorm:"column:material_id;not null;index"`
}

// Point 点位
type Point struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"index;not null"`
	SizeID     int    `gorm:"column:size_id;not null;index"`
	Index      int    `gorm:"not null"`
	UpperLimit float64
	LowerLimit float64
	Norminal   float64
}

// NotValid 校验数据有效性
func (p *Point) NotValid(v float64) bool {
	return p.Norminal > 0 && v > p.Norminal*100
}

// PointValue 点位值
type PointValue struct {
	PointID     int     `gorm:"column:point_id;not null;index"`
	ProductUUID string  `gorm:"column:product_uuid;not null;index"`
	V           float64 `gorm:"column:v;not null"`
}

// File 存储已加载数据的文件路径
type File struct {
	ID           int
	Path         string
	MaterialID   int
	TotalRows    int
	FinishedRows int
	Finished     bool `gorm:"default:false"`
	FileDate     time.Time
}

func init() {
	var err error
	var dbUrl string

	if dbdns := os.Getenv("DB_DNS"); dbdns != "" {
		dbUrl = dbdns
	} else if conf.GetEnv() == "TEST" {
		dbUrl = conf.DBdnstest
	} else {
		dbUrl = conf.DBdns
	}

	reconnectLimit := 5
	for {
		DB, err = gorm.Open("mysql", dbUrl)
		if err != nil && reconnectLimit > 0 {
			reconnectLimit--
			time.Sleep(time.Duration(5-reconnectLimit) * 2 * time.Second)
			fmt.Println("try to reconnect db again ...")
			continue
		}
		break
	}

	if os.Getenv("ENV") == "production" {
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

	generateRootUser()
	generateDefaultConfig()
}

func generateDefaultConfig() {
	if conf.GetEnv() != "TEST" {
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
	if conf.GetEnv() == "TEST" {
		DB.Exec("DELETE FROM system_configs WHERE 1 = 1")
	}
	t := time.Now()
	var sql = `
	INSERT INTO system_configs (system_configs.key, system_configs.value, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	`
	DB.Exec(sql, "ftp_password", "abc1234.", t, t)
	DB.Exec(sql, "ftp_username", "s17695", t, t)
	DB.Exec(sql, "ftp_host", "192.168.2.18", t, t)
	DB.Exec(sql, "ftp_port", "20", t, t)
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
