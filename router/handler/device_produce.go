package handler

import (
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func DeviceProduce() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceToken := c.PostForm("device_token")
		var device orm.Device
		if err := device.GetWithToken(deviceToken); err != nil {
			errormap.SendHttpError(c, err.GetCode(), err, "device")
			return
		}

		qualifiedStr := c.PostForm("qualified")
		qualified, _ := strconv.ParseBool(qualifiedStr)

		attributesStr := c.PostForm("attributes")
		attribute := make(types.Map)
		kValues := strings.Split(attributesStr, ";")
		for _, item := range kValues {
			sectors := strings.Split(item, ":")
			if len(sectors) < 2 {
				continue
			}

			attribute[sectors[0]] = sectors[1]
		}

		pointValuesStr := c.PostForm("point_values")
		pointValues := make(types.Map)
		kValues = strings.Split(pointValuesStr, ";")
		for _, item := range kValues {
			sectors := strings.Split(item, ":")
			if len(sectors) < 2 {
				continue
			}
			value, err := strconv.ParseFloat(sectors[1], 64)
			if err != nil {
				value = 0
			}
			pointValues[sectors[0]] = value
		}

		var product = orm.Product{
			MaterialID:  device.MaterialID,
			DeviceID:    device.ID,
			Qualified:   qualified,
			Attribute:   attribute,
			PointValues: pointValues,
		}
		if err := orm.Create(&product).Error; err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeCreateObjectError, err, "product")
			return
		}

		c.JSON(http.StatusOK, "ok")
	}
}
