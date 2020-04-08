package logic

import (
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"strconv"
	"time"
)

// ClearUp 清楚过期数据
func ClearUp() {
	log.Println("[ClearUp] Begin clear up worker")
	go func() {
		for {
			select {
			case <-time.After(24 * time.Hour):
				clearUp()
			}
		}
	}()
}

func clearUp() {
	config := orm.GetSystemConfig("cache_days")
	expiredDays, err := strconv.Atoi(config.Value)
	if err != nil {
		expiredDays = 30
	}

	end := time.Now().AddDate(0, 0, -expiredDays)
	orm.DB.Exec("DELETE FROM products WHERE created_at < ?", end)
	orm.DB.Exec("DELETE FROM size_values WHERE created_at < ?", end)
}
