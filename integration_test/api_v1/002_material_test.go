package api_v1

import (
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"testing"
)

const (
	productAttributesGQL = `
	query($materialID: Int!, $versionID: Int) {
	  response: productAttributes(materialID: $materialID, versionID: $versionID) {
	    prefix
	    label
	    token
	  }
	}
	`
)

func TestMaterial(t *testing.T) {
	tester := test.NewTester(t)
	material := test.Data.Material
	device := orm.Device{
		Name:       "process device",
		Remark:     "process device",
		MaterialID: material.ID,
	}
	orm.Create(&device)
	createProducts(material.ID, device.ID, 1)

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
			Name:       "process device 1",
			Remark:     "process device 1",
			MaterialID: material.ID,
		}
		orm.Create(&device1)
		createProducts(material.ID, device1.ID, 3)
		device2 := orm.Device{
			Name:       "process device 2",
			Remark:     "process device 2",
			MaterialID: material.ID,
		}
		orm.Create(&device2)
		createProducts(material.ID, device2.ID, 2)

		tester.API1(analyzeMaterialGQL, test.Object{
			"input": test.Object{
				"materialID": material.ID,
				"xAxis":      "Device",
				"yAxis":      "Yield",
				//"groupBy":        "Device",
				//"attributeXAxis": "NO.",
				// "duration": []string{"2020-06-22T00:00:00Z", "2020-06-26T00:00:00Z"},
				//"limit": 2,
				"sort": "DESC",
			},
		}).GQLObject().Path("$.data.response")
	})

	// Test material yield top
	t.Run("TEST_MATERIAL_YIELD_TOP", func(t *testing.T) {
		m1 := orm.Material{Name: "m1"}
		orm.Create(&m1)
		createProducts(m1.ID, 0, 2)
		m2 := orm.Material{Name: "m2"}
		orm.Create(&m2)
		createProducts(m2.ID, 0, 3)
		m3 := orm.Material{Name: "m3"}
		orm.Create(&m3)
		createProducts(m3.ID, 0, 4)

		tester.API1(materialYieldTopGQL, test.Object{
			"duration": []string{},
			"limit":    4,
		}).GQLObject().Path("$.data.response")
	})

	t.Run("TEST_ProductAttributes", func(t *testing.T) {
		materialID := fakeForTestProductAttributes()
		ret := tester.API1(productAttributesGQL, test.Object{"materialID": materialID}).GQLObject().Path("$.data.response")
		ret.Array().Length().Equal(2)
	})
}

func fakeForTestProductAttributes() uint {
	itemsMap := make(types.Map)
	items := []orm.BarCodeItem{
		{
			Label:      "LabelForItem1",
			Key:        "AKeyForItem",
			IndexRange: []int{},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "LabelForItem2",
			Key:        "BKeyForItem",
			IndexRange: []int{},
			Type:       model.BarCodeItemTypeDatetime.String(),
			DayCode:    []string{"1", "V"},
			MonthCode:  []string{"1", "A"},
		},
	}
	itemsMap["items"] = items
	var rule = orm.BarCodeRule{
		CodeLength: 0,
		Name:       "fake_rule_name",
		Remark:     "fake_rule_remark",
		UserID:     test.Data.Admin.ID,
		Items:      itemsMap,
	}
	orm.Create(&rule)

	material := orm.Material{
		Name: "Material_For_TEST_ProductAttributes",
	}
	orm.Create(&material)

	version := orm.MaterialVersion{
		Version:    "Version_For_TEST_ProductAttributes",
		MaterialID: material.ID,
		Active:     true,
		UserID:     test.Data.Admin.ID,
	}
	orm.Create(&version)

	template := orm.DecodeTemplate{
		MaterialID:        material.ID,
		MaterialVersionID: version.ID,
		UserID:            test.Data.Admin.ID,
		BarCodeRuleID:     rule.ID,
		ProductColumns:    make(types.Map),
	}
	orm.Create(&template)
	return material.ID
}

func createProducts(materialID, deviceID uint, rate int) {
	orm.DB.LogMode(false)
	for i := 0; i < 60; i++ {
		product := &orm.Product{
			MaterialID: materialID,
			DeviceID:   deviceID,
			Attribute:  types.Map{"NO.": i % 10},
		}

		if i%rate == 0 {
			product.Qualified = true
		}
		orm.Create(product)
	}
	orm.DB.LogMode(true)
}
