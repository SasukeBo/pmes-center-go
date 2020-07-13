package handler

import (
	"encoding/base64"
	"errors"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Active() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("active_token")
		if err := active(token); err != nil {
			c.Header("content-type", "application/json")
			c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}

		c.Header("content-type", "application/json")
		c.AbortWithStatusJSON(http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "active",
		})
	}
}

var durations = map[string]int{
	"week":      7,
	"week2":     14,
	"month":     30,
	"month2":    60,
	"month3":    90,
	"year":      365,
	"unlimited": 0,
}

func active(token string) error {
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

	config := orm.SystemConfig{}
	config.GetConfig("expired_at")
	config.Key = "expired_at"
	config.Value = expiredValue

	if err := orm.DB.Model(config).Save(config).Error; err != nil {
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
