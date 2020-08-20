package orm

import (
	"encoding/json"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
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

const pageLength = 1000
const pDuration = 24 * time.Hour

// 存储id连续的products
// 分页缓存，10000的整数倍
func cacheProducts(products []Product) {
	t1 := time.Now()
	_, _ = cache.Pipelined(func(pip redis.Pipeliner) error {
		for _, p := range products {
			key := getPK(int(p.ID))
			err := cache.SetWithPip(pip, key, p, pDuration)
			if err != nil {
				log.Errorln(err)
			}
		}
		return nil
	})
	util.DebugTime(t1, "cache product:")
}

func getPK(id int) string {
	return fmt.Sprintf("product-%v", id)
}

type PipGet struct {
	Get *redis.StringCmd
	ID  int
}

func FetchProducts(ids []int, query *gorm.DB) []Product {
	log.Info("start fetch length: %v", len(ids))
	var results []Product
	var unHits []int

	t1 := time.Now()
	var pgs []PipGet
	_, _ = cache.Pipelined(func(pip redis.Pipeliner) error {
		for _, id := range ids {
			pg := PipGet{
				Get: pip.Get(cache.Ctx(), getPK(id)),
				ID:  id,
			}
			pgs = append(pgs, pg)
		}

		return nil
	})
	for _, pg := range pgs {
		var p Product
		if pg.Get.Err() == redis.Nil {
			unHits = append(unHits, pg.ID)
			continue
		}
		if err := json.Unmarshal([]byte(pg.Get.Val()), &p); err != nil {
			unHits = append(unHits, pg.ID)
			continue
		}
		results = append(results, p)
	}
	_ = util.DebugTime(t1, "Pipe get cache spend")

	if unHitCount := len(unHits); unHitCount > 0 {
		log.Info("total %v, unHits length = %v", len(ids), unHitCount)
		var rest []Product
		DB.LogMode(false)
		defer DB.LogMode(true)
		if err := Model(&Product{}).Where("id in (?)", unHits).Find(&rest).Error; err != nil {
			log.Errorln(err)
			if err := query.Find(&results).Error; err == nil {
				go cacheProducts(results)
				return results
			}
		}
		results = append(results, rest...)
		go cacheProducts(rest)
	}

	log.Infoln("len(results)", len(results))
	return results
}
