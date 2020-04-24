package logic

import (
	"encoding/base64"
	"errors"
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"strconv"
	"time"
)

var durations = map[string]int{
	"week":      7,
	"week2":     14,
	"month":     30,
	"month2":    60,
	"month3":    90,
	"year":      365,
	"unlimited": 0,
}

// ClearUp 清除过期数据
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

func Active(token string) error {
	t := time.Now()
	// 加上时间戳，精确到小时，一小时内有效，必须是UTC时间
	date := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	activeDuration := -1
	for d, i := range durations {
		gt := base64.StdEncoding.EncodeToString([]byte(date.String() + d))
		if gt == token {
			activeDuration = i
			break
		}
	}

	if activeDuration == -1 {
		return errors.New("illegal active token")
	}

	var expiredValue string
	if activeDuration == 0 {
		expiredValue = "unlimited"
	} else {
		t := time.Now().AddDate(0, 0, activeDuration)
		expiredValue = t.Format(time.RFC3339)
	}

	expiredConfig := orm.GetSystemConfig("expired_at")
	if expiredConfig == nil {
		expiredConfig = &orm.SystemConfig{Key: "expired_at"}
	}
	expiredConfig.Value = expiredValue

	if err := orm.DB.Model(expiredConfig).Save(expiredConfig).Error; err != nil {
		return err
	}

	return nil
}

func genActiveToken(duration string) string {
	t := time.Now().UTC()
	// 加上时间戳，精确到小时，一小时内有效，必须是UTC时间
	date := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	gt := base64.StdEncoding.EncodeToString([]byte(date.String() + duration))
	return gt
}
