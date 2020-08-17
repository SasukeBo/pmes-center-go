package handler

import (
	"encoding/json"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type response struct {
	DeviceToken string `json:"device_token"`
	PointValues string `json:"point_values"`
	Attributes  string `json:"attributes"`
	Qualified   int    `json:"qualified"`
	BarCode     string `json:"bar_code"`
}

var conn *gorm.DB

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
		if err := device.GetWithToken(deviceToken, conn); err != nil {
			errormap.SendHttpError(c, err.GetCode(), err, "device")
			return
		}
		ip := c.Request.Header.Get("X-Real-IP")
		if device.IP != ip {
			device.IP = ip
			_ = conn.Save(&device)
		}

		var record orm.ImportRecord
		if err := record.GetDeviceRealtimeRecord(&device, conn); err != nil {
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

		rule := device.GetCurrentTemplateDecodeRule(conn)
		barCode := strings.TrimSpace(form.BarCode)
		if rule != nil {
			decoder := logic.NewBarCodeDecoder(rule)
			attribute, statusCode = decoder.Decode(barCode)
		} else {
			attribute = make(types.Map)
		}

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
			BarCode:           barCode,
			BarCodeStatus:     statusCode,
		}
		if err := conn.Create(&product).Error; err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeCreateObjectError, err, "product")
			return
		}
		record.Increase(1, 1, qualified, conn)
		c.JSON(http.StatusOK, "ok")
	}
}

func init() {
	conn = orm.NewConnection()
	conn.LogMode(false)
}
