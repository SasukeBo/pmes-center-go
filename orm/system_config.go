package orm

import "github.com/jinzhu/gorm"

// SystemConfig 系统设置表
type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"unique_index"`
	Value string
}

func SetIfNotExist(key, value string) {
	var s SystemConfig
	if err := DB.Model(&SystemConfig{}).Where("`system_configs`.`key` = ?", key).First(&s).Error; err != nil {
		s = SystemConfig{
			Key:   key,
			Value: value,
		}

		DB.Create(&s)
	}
}

func (s *SystemConfig) GetConfig(key string) error {
	return DB.Model(s).Where("`system_configs`.`key` = ?", key).First(s).Error
}
