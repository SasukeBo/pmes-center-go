package admin

import (
	test "github.com/SasukeBo/ftpviewer/intergration_test"
	"testing"
)

func TestDecodeTemplate(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)

	t.Run("Test create decode_template", func(t *testing.T) {
		ret := tester.API1Admin(saveDecodeTemplateGQL, test.Object{
			"input": test.Object{
				"name":                 "test decode template",
				"materialID":           test.Data.Material.ID,
				"description":          "test decode template description",
				"dataRowIndex":         15,
				"createdAtColumnIndex": 3,
				"productColumns": []test.Object{
					{"name": "No.", "index": 1, "type": "Integer"},
					{"name": "日期", "index": 2, "type": "Datetime"},
					{"name": "线体号", "index": 3, "type": "String"},
					{"name": "精度", "index": 4, "type": "Float"},
				},
				"pointColumns": test.Object{
					"FAI_G5": 5,
					"FAI_G6": 6,
					"FAI_G7": 7,
					"FAI_G8": 8,
				},
				"default": true,
			},
		}).GQLObject().Path("$.data.response").Object()
		productColumns := ret.Value("productColumns").Array()
		productColumns.Length().Equal(4)
		productColumns.First().Object().Value("name").Equal("No.")

		pointColumns := ret.Value("pointColumns").Object()
		pointColumns.Value("FAI_G5").Equal(5)
	})

}
