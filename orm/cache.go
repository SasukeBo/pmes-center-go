package orm

import (
	"fmt"
)

var cache map[string]interface{}

func init() {
	cache = make(map[string]interface{})
}

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

// GetMaterialWithIDCache _
func GetMaterialWithIDCache(materialID string) *Material {
	cacheKey := fmt.Sprintf("cache_material_%s", materialID)
	if m, ok := cache[cacheKey].(Material); ok {
		return &m
	}

	var m Material
	if err := DB.Where("name = ?", materialID).First(&m).Error; err != nil {
		return nil
	}

	cache[cacheKey] = m
	return &m
}

// CacheMaterial _
func CacheMaterial(m Material) {
	cache[fmt.Sprintf("cache_material_%s", m.Name)] = m
}

// GetDeviceWithNameCache _
func GetDeviceWithNameCache(dn string) *Device {
	cacheKey := fmt.Sprintf("cache_device_%s", dn)
	if d, ok := cache[cacheKey].(Device); ok {
		return &d
	}

	var d Device
	if err := DB.Where("name = ?", dn).First(&d).Error; err != nil {
		return nil
	}

	cache[cacheKey] = d
	return &d
}

// CacheDevice _
func CacheDevice(d Device) {
	cache[fmt.Sprintf("cache_device_%s", d.Name)] = d
}

// GetSizeWithMaterialIDSizeNameCache _
func GetSizeWithMaterialIDSizeNameCache(sn, mn string) *Size {
	cacheKey := fmt.Sprintf("cache_size_%s_%s", mn, sn)
	if s, ok := cache[cacheKey].(Size); ok {
		return &s
	}

	var s Size
	if err := DB.Where("name = ? AND material_id = ?", sn, mn).First(&s).Error; err != nil {
		return nil
	}

	cache[cacheKey] = s
	return &s
}

// CacheSize _
func CacheSize(s Size) {
	cache[fmt.Sprintf("cache_size_%s_%s", s.MaterialID, s.Name)] = s
}
