package orm

// 料号的生产设备
// 生产设备的创建方式有两种
// 1.通过数据文件名称解析
// 2.通过后台手动创建

import (
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

const deviceCacheKey = "cache_device_%v_%v"

type Device struct {
	gorm.Model
	UUID           string `gorm:"column:uuid;unique_index;not null"`
	Name           string `gorm:"not null"`                                    // 用于存储用户指定的设备名称，不指定时，默认为Remark的值
	Remark         string `gorm:"not null;unique_index:uidx_name_material_id"` // 用于存储从数据文件解析出的名称
	IP             string `gorm:"column:ip;"`
	MaterialID     uint   `gorm:"column:material_id;not null;unique_index:uidx_name_material_id"` // 同一料号下的设备remark不可重复
	DeviceSupplier string
	IsRealtime     bool `gorm:"default:false;not null"`
	Address        string
}

/*	callbacks
--------------------------------------------------------------------------------------------------------------------- */
func (d *Device) BeforeCreate() error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	d.UUID = uid.String()
	return nil
}

// 清除缓存
func (d *Device) AfterUpdate() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(deviceCacheKey, "token", d.UUID))
	return nil
}
func (d *Device) AfterDelete() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(deviceCacheKey, "token", d.UUID))
	return nil
}
func (d *Device) AfterSave() error {
	_ = cache.FlushCacheWithKey(fmt.Sprintf(deviceCacheKey, "token", d.UUID))
	return nil
}

// 清除缓存

/*	functions
--------------------------------------------------------------------------------------------------------------------- */
func (d *Device) GetWithToken(token string) *errormap.Error {
	cacheKey := fmt.Sprintf(deviceCacheKey, "token", token)
	cacheValue := cache.Get(cacheKey)
	if cacheValue != nil {
		device, ok := cacheValue.(Device)
		if ok {
			if err := copier.Copy(d, &device); err == nil {
				log.Info("get device from cache")
				return nil
			}
		}
	}

	if err := DB.Model(d).Where("uuid = ?", token).First(d).Error; err != nil {
		return handleError(err, "token", token)
	}
	_ = cache.Set(cacheKey, *d)
	return nil
}

func (d *Device) GetWithName(name string) *errormap.Error {
	if err := DB.Model(d).Where("name = ?", name).First(d).Error; err != nil {
		return handleError(err, "name", name)
	}

	return nil
}

func (d *Device) Get(id uint) *errormap.Error {
	if err := DB.Model(d).Where("id = ?", id).First(d).Error; err != nil {
		return handleError(err, "id", id)
	}

	return nil
}

func (d *Device) CreateIfNotExist(materialID uint, remark string) error {
	DB.Model(d).Where("material_id = ? AND remark = ?", materialID, remark).First(d)
	if d.ID == 0 {
		d.Name = remark
		d.MaterialID = materialID
		d.Remark = remark
		err := DB.Create(d).Error
		return err
	}

	return nil
}

func (d *Device) genTemplateDecodeRuleKey() string {
	return fmt.Sprintf("device_current_version_template_rule_key_%v_%s", d.ID, util.NowDateStr())
}

func (d *Device) GetCurrentTemplateDecodeRule() *BarCodeRule {
	key := d.genTemplateDecodeRuleKey()
	value := cache.Get(key)
	if value != nil {
		rule, ok := value.(*BarCodeRule)
		if ok {
			_ = cache.Set(key, rule)
			return rule
		}
	}

	var template DecodeTemplate
	query := Model(&DecodeTemplate{}).Joins("JOIN material_versions ON decode_templates.material_version_id = material_versions.id")
	query.Where("decode_templates.material_id = ? AND material_versions.active = true", d.MaterialID)
	if err := query.Find(&template).Error; err != nil {
		log.Errorln(err)
		return nil
	}

	var rule BarCodeRule
	if err := rule.Get(template.BarCodeRuleID); err != nil {
		log.Errorln(err)
		return nil
	}

	_ = cache.Set(key, &rule)
	return &rule
}
