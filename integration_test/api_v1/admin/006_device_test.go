package admin

import (
	test "github.com/SasukeBo/ftpviewer/integration_test"
	"github.com/SasukeBo/ftpviewer/orm"
	"testing"
)

func TestDevice(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)

	// Test create device
	t.Run("TEST_CREATE_DEVICE", func(t *testing.T) {
		ret := tester.API1Admin(saveDeviceGQL, test.Object{
			"input": test.Object{
				"name":           "test_device",
				"remark":         "test_remark",
				"ip":             "0.0.0.0",
				"materialID":     test.Data.Material.ID,
				"deviceSupplier": "test device supplier",
				"isRealtime":     true,
				"address":        "test device address",
			},
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("uuid").NotNull()
		ret.Value("name").Equal("test_device")
		ret.Value("ip").Equal("0.0.0.0")
		ret.Path("$.material.id").Equal(test.Data.Material.ID)
	})

	// Test update device
	t.Run("TEST_UPDATE_DEVICE", func(t *testing.T) {
		device := orm.Device{
			Name:           "TEST_UPDATE_DEVICE",
			Remark:         "TEST_UPDATE_DEVICE",
			IP:             "127.0.0.1",
			MaterialID:     test.Data.Material.ID,
			DeviceSupplier: "TEST_UPDATE_DEVICE",
			IsRealtime:     false,
			Address:        "TEST_UPDATE_DEVICE",
		}
		orm.Create(&device)

		ret := tester.API1Admin(saveDeviceGQL, test.Object{
			"input": test.Object{
				"id":         device.ID,
				"name":       "changed name",
				"ip":         "8.8.8.8",
				"remark":     "remark_cannot_be_change",
				"materialID": 0,
				"isRealtime": true,
			},
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("uuid").Equal(device.UUID)
		ret.Value("ip").Equal("8.8.8.8")
		ret.Value("isRealtime").Equal(true)
		ret.Value("name").Equal("changed name")
		ret.Value("remark").NotEqual("remark_cannot_be_change")
		ret.Path("$.material.id").NotEqual(0)
	})

	// Test list devices
	t.Run("TEST_LIST_DEVICES", func(t *testing.T) {
		device1 := orm.Device{
			Name:       "device1",
			Remark:     "device1",
			IP:         "0.0.0.0",
			MaterialID: test.Data.Material.ID,
		}
		orm.Create(&device1)
		device2 := device1
		device2.ID = 0
		device2.Name = "device2"
		device2.Remark = "device2"
		orm.Create(&device2)

		ret := tester.API1Admin(listDevicesGQL, test.Object{
			"pattern":    "device",
			"materialID": test.Data.Material.ID,
			"page":       1,
			"limit":      1,
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("total").Equal(2)
		ret.Value("devices").Array().Length().Equal(1)
	})
}
