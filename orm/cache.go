package orm

import (
	"fmt"
)

var cache map[string]interface{}

// GetSystemConfigCache 获取缓存
func GetSystemConfigCache(key string) *SystemConfig {
	cacheKey := fmt.Sprintf("cache_system_conf_%s", key)
	if conf, ok := cache[cacheKey].(SystemConfig); ok {
		return &conf
	}

	var conf SystemConfig
	if err := DB.Where("system_configs.key = ?", key).Find(&conf).Error; err != nil {
		return nil
	}

	cache[cacheKey] = conf
	return &conf
}

// CacheSystemConfig _
func CacheSystemConfig(m SystemConfig) {
	cache[fmt.Sprintf("cache_system_conf_%s", m.Key)] = m
}

// GetUserWithTokenCache 获取缓存
func GetUserWithTokenCache(token string) *User {
	cacheKey := fmt.Sprintf("cache_user_%s", token)
	if user, ok := cache[cacheKey].(User); ok {
		return &user
	}

	var user User
	if err := DB.Where("access_token = ?", token).Find(&user).Error; err != nil {
		return nil
	}

	cache[cacheKey] = user
	return &user
}

func init() {
	cache = make(map[string]interface{})
}
