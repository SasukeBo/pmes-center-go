package admin

import (
	"github.com/SasukeBo/pmes-data-center/errormap"
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"net/http"
	"testing"
)

var decodeTemplateInput = test.Object{
	"name":                 "test decode template",
	"materialID":           test.Data.Material.ID,
	"description":          "test decode template description",
	"dataRowIndex":         15,
	"createdAtColumnIndex": "B",
	"productColumns": []test.Object{
		{"prefix": "a0", "token": "attr_0", "label": "No.", "index": "A", "type": "Integer"},
		{"prefix": "a1", "token": "attr_1", "label": "日期", "index": "B", "type": "Datetime"},
		{"prefix": "a2", "token": "attr_2", "label": "线体号", "index": "C", "type": "String"},
		{"prefix": "a3", "token": "attr_3", "label": "精度", "index": "D", "type": "Float"},
	},
	"pointColumns": test.Object{
		"FAI_G5": "E",
		"FAI_G6": "F",
		"FAI_G7": "G",
		"FAI_G8": "ABC",
	},
	"default": true,
}

func TestDecodeTemplate(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)

	// Test create decode_template
	t.Run("TEST_CREATE_DECODE_TEMPLATE", func(t *testing.T) {
		ret := tester.API1Admin(saveDecodeTemplateGQL, test.Object{
			"input": decodeTemplateInput,
		}).GQLObject().Path("$.data.response").Object()
		productColumns := ret.Value("productColumns").Array()
		productColumns.Length().Equal(4)
		productColumns.First().Object().Value("name").Equal("No.")
		productColumns.First().Object().Value("index").Equal("A")

		pointColumns := ret.Value("pointColumns").Object()
		pointColumns.Value("FAI_G5").Equal("E")
		ret.Path("$.user.id").Equal(test.Data.Admin.ID)
	})

	// Test save decode_template
	t.Run("TEST_SAVE_DECODE_TEMPLATE", func(t *testing.T) {
		template := orm.DecodeTemplate{
			Name:                 "test decode template",
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			Description:          "test description",
			DataRowIndex:         15,
			CreatedAtColumnIndex: 2,
			Default:              false,
		}
		orm.Create(&template)
		ret := tester.API1Admin(saveDecodeTemplateGQL, test.Object{
			"input": test.Object{
				"id":                   template.ID,
				"name":                 "changed name",
				"materialID":           100,
				"description":          "changed description",
				"dataRowIndex":         15,
				"createdAtColumnIndex": "C",
				"productColumns": []test.Object{
					{"name": "attr_0", "label": "No.", "index": "A", "type": "Integer"},
					{"name": "attr_1", "label": "日期", "index": "B", "type": "Datetime"},
					{"name": "attr_2", "label": "线体号", "index": "C", "type": "String"},
					{"name": "attr_3", "label": "精度", "index": "D", "type": "Float"},
				},
				"pointColumns": test.Object{
					"FAI_G5": "E",
					"FAI_G6": "F",
					"FAI_G7": "G",
					"FAI_G8": "H",
				},
				"default": true,
			},
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("name").Equal("changed name")                // 名称可更改
		ret.Path("$.material.id").Equal(test.Data.Material.ID) // 所属料号不可更改
		ret.Path("$.user.id").Equal(test.Data.User.ID)         // 创建人不变
		ret.Value("description").Equal("changed description")  // 描述可更改
		ret.Value("createdAtColumnIndex").Equal("C")           // 生产日期数据列序号可更改
		ret.Value("default").Equal(true)                       // 可设置为默认模板
		productColumns := ret.Value("productColumns").Array()
		productColumns.Length().Equal(4)
		productColumns.First().Object().Value("name").Equal("No.")
		productColumns.First().Object().Value("index").Equal("A")

		pointColumns := ret.Value("pointColumns").Object()
		pointColumns.Value("FAI_G5").Equal("E")
	})

	// Test list decode_templates
	t.Run("TEST_LIST_DECODE_TEMPLATES", func(t *testing.T) {
		columns := types.Map{
			"attr_0": orm.Column{Label: "No.", Index: 0, Type: "Integer"},
			"attr_1": orm.Column{Label: "日期", Index: 1, Type: "Datetime"},
			"attr_2": orm.Column{Label: "线体号", Index: 2, Type: "String"},
			"attr_3": orm.Column{Label: "精度", Index: 3, Type: "Float"},
		}
		template := orm.DecodeTemplate{
			Name:                 "template1",
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			Description:          "test description",
			DataRowIndex:         15,
			CreatedAtColumnIndex: 1,
			Default:              false,
			ProductColumns:       columns,
			PointColumns: types.Map{
				"FAI_G5": 4,
				"FAI_G6": 5,
				"FAI_G7": 6,
				"FAI_G8": 7,
			},
		}
		orm.Create(&template)
		template.Name = "template2"
		template.ID = 0
		orm.Create(&template)
		template.Name = "template3"
		template.ID = 0
		template.MaterialID = 0
		orm.Create(&template)

		ret := tester.API1Admin(listDecodeTemplateGQL, test.Object{
			"materialID": test.Data.Material.ID,
		}).GQLObject().Path("$.data.response").Array()
		ret.Length().Equal(2)
		ret.First().Object().Value("name").Equal("template1")

		productColumns := ret.First().Object().Value("productColumns").Array()
		productColumns.Length().Equal(4)

		pointColumns := ret.First().Object().Value("pointColumns").Object()
		pointColumns.Value("FAI_G5").Equal("E")
	})

	// Test delete decode template
	t.Run("TEST_DELETE_DECODE_TEMPLATE", func(t *testing.T) {
		tester.SetHeader(errormap.LangHeader, errormap.EN)
		columns := types.Map{
			"attr_0": orm.Column{Label: "No.", Index: 0, Type: "Integer"},
			"attr_1": orm.Column{Label: "日期", Index: 1, Type: "Datetime"},
			"attr_2": orm.Column{Label: "线体号", Index: 2, Type: "String"},
			"attr_3": orm.Column{Label: "精度", Index: 3, Type: "Float"},
		}
		template := orm.DecodeTemplate{
			Name:                 "template1",
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			Description:          "test description",
			DataRowIndex:         15,
			CreatedAtColumnIndex: 2,
			Default:              false,
			ProductColumns:       columns,
			PointColumns: types.Map{
				"FAI_G5": 5,
				"FAI_G6": 6,
				"FAI_G7": 7,
				"FAI_G8": 8,
			},
		}
		orm.Create(&template)
		tester.API1Admin(deleteDecodeTemplateGQL, test.Object{"id": template.ID}).GQLObject().Path("$.data.response").Equal("OK")
		tester.API1Admin(deleteDecodeTemplateGQL, test.Object{"id": 0}).GQLObject().Value("errors").Array().First().Object().Path("$.extensions.code").Equal(http.StatusNotFound)
		template.Default = true
		template.ID = 0
		orm.Create(&template)
		tester.API1Admin(deleteDecodeTemplateGQL, test.Object{"id": template.ID}).GQLObject().Value("errors").Array().First().Object().Path("$.extensions.code").Equal(http.StatusBadRequest)
	})

	// Test change default decode template
	t.Run("TEST_CHANGE_DEFAULT_DECODE_TEMPLATE", func(t *testing.T) {
		template1 := orm.DecodeTemplate{
			Name:                 "template1",
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			Description:          "test description",
			DataRowIndex:         15,
			CreatedAtColumnIndex: 1,
			Default:              true,
			ProductColumns:       types.Map{"columns": []orm.Column{}},
			PointColumns:         types.Map{},
		}
		orm.Create(&template1)
		template2 := template1
		template2.ID = 0
		template2.Name = "template2"
		template2.Default = false
		orm.Create(&template2)

		tester.API1Admin(changeDefaultTemplateGQL, test.Object{
			"id":        template2.ID,
			"isDefault": true,
		}).GQLObject().Path("$.data.response").Equal("OK")
	})
}
