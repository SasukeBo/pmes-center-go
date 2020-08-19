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

const pageLength = 10000

// 存储id连续的products
// 分页缓存，10000的整数倍
func cacheProducts(products []Product) {
	if len(products) > pageLength {
		products = products[0:pageLength]
	}
	key := getPageKey(int(products[0].ID))
	fmt.Printf("set page key=%s, product[0]=%v, product[len-1]=%v\n", key, products[0].ID, products[len(products)-1].ID)
	err := cache.Set(key, products, 24*time.Hour)
	if err != nil {
		fmt.Println(err)
	}
}

type productPageMap map[int][]Product

func getPageKey(id int) string {
	return fmt.Sprintf("product-%v", id/pageLength)
}

func getPageNum(id int) int {
	return id / pageLength
}

func getPageIndex(id int) int {
	return id % pageLength
}

func reduceGetIndex(page []Product, id int) *Product {
	pi := getPageIndex(id)
	if pi >= len(page) {
		pi = len(page) - 1
	}

	if hit := page[pi]; int(hit.ID) == id {
		return &hit
	} else {
		if int(hit.ID) < id {
			return nil
		}

		var begin = 0
		var end = len(page) - 1
		for {
			mid := (begin + end) / 2
			if guess := page[mid]; int(guess.ID) == id {
				return &guess
			} else if int(guess.ID) > id {
				end = mid
			} else if int(guess.ID) < id {
				begin = mid
			}
		}
	}
}

func FetchProducts(ids []int) []Product {
	log.Info("start fetch length: %v", len(ids))
	var results []Product
	var unHits []int
	var pageMap = make(productPageMap)
	t1 := time.Now()
	for _, id := range ids {
		pn := getPageNum(id)
		var hit Product
		if page, ok := pageMap[pn]; ok {
			guess := reduceGetIndex(page, id)
			if guess == nil {
				unHits = append(unHits, id)
				continue
			}
			hit = *guess
		} else {
			key := getPageKey(id)
			if err := cache.Scan(key, &page); err != nil {
				unHits = append(unHits, id)
				continue
			} else {
				guess := reduceGetIndex(page, id)
				if guess == nil {
					unHits = append(unHits, id)
					continue
				}
				hit = *guess
			}
			pageMap[pn] = page
		}
		if int(hit.ID) != id {
			log.Error("id not match: %v %v", hit.ID, id)
		}
		results = append(results, hit)
	}
	util.DebugTime(t1, "time during cache loop")

	if len(unHits) > 0 {
		go CacheProductPage(unHits)
		//var rest []Product
		//if err := Model(&Product{}).Where("id in (?)", unHits).Find(&rest).Error; err == nil {
		//	results = append(results, rest...)
		//}
	}
	return results
}

func CacheProductPage(ids []int) {
	var currentPageNum = 0
	for _, id := range ids {
		pn := getPageNum(id)
		if pn > currentPageNum {
			go func(thisPN int) {
				var page []Product
				query := Model(&Product{}).Where("id >= ? AND id < ?", thisPN*pageLength, (thisPN+1)*pageLength)
				if err := query.Limit(pageLength).Find(&page).Error; err == nil {
					cacheProducts(page)
				}
			}(pn)
			currentPageNum = pn
		}
	}
}
