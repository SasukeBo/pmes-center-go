package handler

import (
	"encoding/json"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func DeviceProduce1() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceToken := c.PostForm("device_token")
		var device orm.Device
		if err := device.GetWithToken(deviceToken); err != nil {
			errormap.SendHttpError(c, err.GetCode(), err, "device")
			return
		}

		qualifiedStr := c.PostForm("qualified")
		qualifiedInt, _ := strconv.ParseInt(qualifiedStr, 10, 64)
		var qualified bool
		if qualifiedInt == 1 {
			qualified = true
		}

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

type response struct {
	DeviceToken string `json:"device_token"`
	PointValues string `json:"point_values"`
	Attributes  string `json:"attributes"`
	Qualified   int    `json:"qualified"`
	BarCode     string `json:"bar_code"`
}

func DeviceProduce() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := ioutil.ReadAll(c.Request.Body)
		var form response
		if err := json.Unmarshal(body, &form); err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeBadRequestParams, err)
			return
		}

		deviceToken := form.DeviceToken
		var device orm.Device
		if err := device.GetWithToken(deviceToken); err != nil {
			errormap.SendHttpError(c, err.GetCode(), err, "device")
			return
		}

		var record orm.ImportRecord
		if err := record.GetDeviceRealtimeRecord(&device); err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeInternalError, err, "record")
		}

		qualifiedInt := form.Qualified
		var qualified bool
		if qualifiedInt == 1 {
			qualified = true
		}

		//attributesStr := form.Attributes
		var attribute types.Map
		var statusCode = 1

		rule := device.GetCurrentTemplateDecodeRule()
		if rule != nil {
			decoder := logic.NewBarCodeDecoder(rule)
			attribute, statusCode = decoder.Decode(form.BarCode)
		} else {
			attribute = make(types.Map)
		}

		// TODO: deprecate
		//kValues := strings.Split(attributesStr, ";")
		//for _, item := range kValues {
		//	sectors := strings.Split(item, ":")
		//	if len(sectors) < 2 {
		//		continue
		//	}
		//
		//	attribute[sectors[0]] = sectors[1]
		//}

		pointValuesStr := form.PointValues
		pointValues := make(types.Map)
		kValues := strings.Split(pointValuesStr, ";")
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
			MaterialID:        device.MaterialID,
			DeviceID:          device.ID,
			Qualified:         qualified,
			Attribute:         attribute,
			PointValues:       pointValues,
			ImportRecordID:    record.ID,
			MaterialVersionID: record.MaterialVersionID,
			BarCode:           form.BarCode,
			BarCodeStatus:     statusCode,
		}
		if err := orm.Create(&product).Error; err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeCreateObjectError, err, "product")
			return
		}
		record.Increase(1, 1, qualified)
		c.JSON(http.StatusOK, "ok")
	}
}
