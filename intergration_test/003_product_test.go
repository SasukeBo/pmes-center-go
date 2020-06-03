package test

import (
	"github.com/SasukeBo/ftpviewer/orm"
	"testing"
	"time"
)

func TestProduct(t *testing.T) {
	tester := newTester(t)
	material := &orm.Material{
		Name:          "test_material",
		CustomerCode:  "test_material_customer_code",
		ProjectRemark: "test_material_project_remark",
	}
	orm.Create(material)
	device := &orm.Device{
		Name:           "test_device",
		Remark:         "test_device",
		MaterialID:     material.ID,
		DeviceSupplier: "test_device_supplier",
	}
	orm.Create(device)
	importRecord := &orm.ImportRecord{
		MaterialID: material.ID,
		DeviceID:   device.ID,
		ImportType: orm.ImportRecordTypeSystem,
	}
	orm.Create(importRecord)
	insertProduct(
		importRecord.ID, material.ID, device.ID, true,
		`{"NO.": 2, "日期": "2020-04-23T15:09:46Z", "模号": "A", "班别": "4", "冶具号": "16", "线体号": "1", "2D条码号": "FTTA17301JK2760AT116A417417Z"}`,
		`{"FAI3_G7": 5.306, "FAI3_G8": 5.3, "FAI4_G1": 4.212, "FAI4_G2": 4.214, "FAI4_G3": 4.215}`,
	)
	insertProduct(
		importRecord.ID, material.ID, device.ID, false,
		`{"NO.": 6, "日期": "2020-04-23T15:09:56Z", "模号": "A", "班别": "4", "冶具号": "01", "线体号": "1", "2D条码号": "FTTA17301G42760AE101A417417Z"}`,
		`{"FAI3_G7": 5.306, "FAI3_G8": 5.305, "FAI4_G1": 4.221, "FAI4_G2": 4.216, "FAI4_G3": 4.212}`,
	)

	t.Run("Search with normal option", func(t *testing.T) {
		ret := tester.API1(productScrollFetchGQL, object{
			"input": object{
				"materialID":     material.ID,
				"importRecordID": importRecord.ID,
				"deviceID":       device.ID,
				"attributes":     object{},
			},
			"limit":  100,
			"offset": 0,
		}).GQLObject().Path("$.data.response")
		ret.Object().Value("total").Equal(2)
	})

	t.Run("Search with attributes", func(t *testing.T) {
		ret := tester.API1(productScrollFetchGQL, object{
			"input": object{
				"materialID": material.ID,
				"attributes": object{
					"NO.": 6,
					"模号":  "A",
				},
			},
			"limit":  100,
			"offset": 0,
		}).GQLObject().Path("$.data.response")
		ret.Object().Value("total").Equal(1)
		ret.Object().Value("data").Array().First().Object().Value("attribute").Object().Value("冶具号").Equal("01")
	})
}

func insertProduct(importRecordID, materialID, deviceID uint, qualified bool, attribute, pointValues string) {
	orm.Exec(`
	INSERT INTO products (import_record_id, material_id, device_id, qualified, created_at, products.attribute, point_values)
		values(?, ?, ?, ?, ?, ?, ?)	
	`, importRecordID, materialID, deviceID, qualified, time.Now(), attribute, pointValues)
}
