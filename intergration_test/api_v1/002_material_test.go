package api_v1

import (
	test "github.com/SasukeBo/ftpviewer/intergration_test"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"testing"
)

func TestMaterial(t *testing.T) {
	tester := test.NewTester(t)
	material := test.Data.Material
	device := orm.Device{
		Name:       "test device",
		Remark:     "test device",
		MaterialID: material.ID,
	}
	orm.Create(&device)
	createProducts(material.ID, device.ID)

	// Test get list of materials
	t.Run("TEST_GET_LIST_OF_MATERIALS", func(t *testing.T) {
		res := tester.API1(materialsGQL, test.Object{
			"page":  1,
			"limit": 10,
		}).GQLObject().Path("$.data.response").Object()
		res.Value("total").Equal(1)
		materialRes := res.Value("materials").Array().First().Object()
		materialRes.Value("name").Equal(material.Name)
		materialRes.Value("ok").Equal(25)
		materialRes.Value("ng").Equal(25)
	})

	// Test get material by id
	t.Run("TEST_GET_MATERIAL_BY_ID", func(t *testing.T) {
		res := tester.API1(materialGQL, test.Object{"id": material.ID}).GQLObject().Path("$.data.response").Object()
		res.Value("name").Equal(material.Name)
		res.Value("ok").Equal(25)
		res.Value("ng").Equal(25)
	})

	// Test analyze material
	t.Run("TEST_ANALYZE_MATERIAL", func(t *testing.T) {
		device1 := orm.Device{
			Name:       "test device 1",
			Remark:     "test device 1",
			MaterialID: material.ID,
		}
		orm.Create(&device1)
		createProducts(material.ID, device1.ID)
		device2 := orm.Device{
			Name:       "test device 2",
			Remark:     "test device 2",
			MaterialID: material.ID,
		}
		orm.Create(&device2)
		createProducts(material.ID, device2.ID)

		tester.API1(analyzeMaterialGQL, test.Object{
			"input": test.Object{
				"materialID":     material.ID,
				"xAxis":          "Attribute",
				"yAxis":          "Yield",
				"groupBy":        "Device",
				"attributeXAxis": "NO.",
				// "duration": []string{"2020-06-22T00:00:00Z", "2020-06-26T00:00:00Z"},
				//"limit": 2,
				"sort":  "DESC",
			},
		}).GQLObject().Path("$.data.response")
	})
}

func createProducts(materialID, deviceID uint) {
	orm.DB.LogMode(false)
	for i := 0; i < 50; i++ {
		product := &orm.Product{
			MaterialID: materialID,
			DeviceID:   deviceID,
			Attribute:  types.Map{"NO.": i % 10},
		}

		if i%2 == 1 {
			product.Qualified = true
		}
		orm.Create(product)
	}
	orm.DB.LogMode(true)
}
