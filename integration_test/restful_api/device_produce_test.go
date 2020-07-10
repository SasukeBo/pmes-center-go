package restful_api

import (
	test "github.com/SasukeBo/ftpviewer/integration_test"
	"github.com/SasukeBo/ftpviewer/orm"
	"net/http"
	"testing"
)

func TestDeviceProducing(t *testing.T) {
	tester := test.NewTester(t)
	device := orm.Device{
		Name:       "test device",
		Remark:     "test device",
		MaterialID: test.Data.Material.ID,
		IsRealtime: true,
	}
	orm.Create(&device)

	pointValues := "FAI95:9.664;FAI96:0.02;FAI97:0.029;FAI98:8.618;FAI99:0.026;FAI100:9.792;FAI96-X:7.689"
	attributes := "NO.:2;日期:2020-06-15T02:53:14Z;模号:;班别:;冶具号:;线体号:QJ;2D条码号:"

	tester.POST("/produce", test.Object{
		"device_token": device.UUID,
		"qualified":    true,
		"point_values": pointValues,
		"attributes":   attributes,
	}).Expect().Status(http.StatusOK).JSON().String().Equal("ok")
}

func BenchmarkDeviceProducing(b *testing.B) {
	tester := test.NewTester(b)
	pointValues := "FAI95:9.664;FAI96:0.02;FAI97:0.029;FAI98:8.618;FAI99:0.026;FAI100:9.792;FAI96-X:7.689"
	attributes := "NO.:2;日期:2020-06-15T02:53:14Z;模号:;班别:;冶具号:;线体号:QJ;2D条码号:"

	for i := 0; i < b.N; i++ {
		tester.POST("/produce", test.Object{
			"device_token": test.Data.Device.UUID,
			"qualified":    true,
			"point_values": pointValues,
			"attributes":   attributes,
		}).Expect().Status(http.StatusOK).JSON().String().Equal("ok")
	}
}
