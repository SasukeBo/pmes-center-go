package orm

import (
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/SasukeBo/pmes-data-center/util"
	"time"
)

// Product 产品表
type Product struct {
	ID                uint      `gorm:"column:id;primary_key"`
	ImportRecordID    uint      `gorm:"COMMENT:'导入记录ID';column:import_record_id;not null;index"`
	MaterialVersionID uint      `gorm:"COMMENT:'料号版本ID';index"`
	MaterialID        uint      `gorm:"COMMENT:'料号ID';column:material_id;not null;index"`
	DeviceID          uint      `gorm:"COMMENT:'检测设备ID';column:device_id;not null;index"`
	Qualified         bool      `gorm:"COMMENT:'产品尺寸是否合格';column:qualified;default:false"`
	BarCode           string    `gorm:"COMMENT:'识别条码';column:bar_code;"`
	BarCodeStatus     int       `gorm:"COMMENT:'条码解析状态';column:bar_code_status;default:1"`
	CreatedAt         time.Time `gorm:"COMMENT:'产品检测时间';index"` // 检测时间
	Attribute         types.Map `gorm:"COMMENT:'产品属性值集合';type:JSON;not null"`
	PointValues       types.Map `gorm:"COMMENT:'产品点位检测值集合';type:JSON;not null"`
}

func pk(id int) string {
	return fmt.Sprintf("product_%d", id)
}

func SetProducts(products []Product) {
	for _, p := range products {
		_ = cache.Set(pk(int(p.ID)), p, 24*time.Hour)
	}
}

func FetchProducts(ids []int) []Product {
	t1 := time.Now()
	log.Info("start fetch length: %v", len(ids))
	var results []Product
	var unHits []int
	for _, id := range ids {
		var hit Product
		if err := cache.Scan(pk(id), &hit); err != nil {
			unHits = append(unHits, id)
			continue
		}
		results = append(results, hit)
	}
	t2 := util.DebugTime(t1, "read redis")
	log.Info("fetch length: %v", len(results))
	log.Info("un hit length: %v", len(unHits))

	if len(unHits) > 0 {
		var rest []Product
		if err := Model(&Product{}).Where("id in (?)", unHits).Find(&rest).Error; err == nil {
			results = append(results, rest...)
			go SetProducts(rest)
		}
	}
	_ = util.DebugTime(t2, "read db")

	log.Info("total length: %v", len(results))
	return results
}
