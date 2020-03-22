package orm

// GetSystemConfig 获取缓存
func GetSystemConfig(key string) *SystemConfig {
	var conf SystemConfig
	if err := DB.Where("system_configs.key = ?", key).Find(&conf).Error; err != nil {
		return nil
	}
	return &conf
}

// GetUserWithToken 获取缓存
func GetUserWithToken(token string) *User {
	var user User
	if err := DB.Where("access_token = ?", token).Find(&user).Error; err != nil {
		return nil
	}
	return &user
}

// GetMaterialWithID _
func GetMaterialWithID(materialID string) *Material {
	var m Material
	if err := DB.Where("name = ?", materialID).First(&m).Error; err != nil {
		return nil
	}
	return &m
}

// GetDeviceWithName _
func GetDeviceWithName(dn string) *Device {
	var d Device
	if err := DB.Where("name = ?", dn).First(&d).Error; err != nil {
		return nil
	}
	return &d
}

// GetSizeWithMaterialIDSizeName _
func GetSizeWithMaterialIDSizeName(sn, mn string) *Size {
	var s Size
	if err := DB.Where("name = ? AND material_id = ?", sn, mn).First(&s).Error; err != nil {
		return nil
	}
	return &s
}

func GetFileListWithPath(path string) *FileList {
	var fl FileList
	if err := DB.Where("path = ?", path).First(&fl).Error; err != nil {
		return nil
	}
	return &fl
}
