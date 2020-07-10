package api_v1

import (
	"fmt"
	test "github.com/SasukeBo/ftpviewer/integration_test"
	"testing"
)

func TestSize(t *testing.T) {
	tester := test.NewTester(t)

	// test SizeUnYieldTop
	t.Run("TEST_SizeUnYieldTop", func(t *testing.T) {
		materialID := generateMaterialData(tester)
		fmt.Println(materialID)
	})
}

// 调用接口创建料号，并且导入对应数据
// createMaterialGQL 调用说明参见 integration/api_v1/admin/002_material_test.go - TestMaterial
func generateMaterialData(tester *test.Tester) uint {
	var createMaterialGQL = `
	mutation($input: MaterialCreateInput!) {
		response: addMaterial(input: $input) {
			id
			name
			customerCode
			projectRemark
		}
	}
	`

	test.Login(test.AdminAccount, test.AdminPasswd, true)
	ret := tester.API1Admin(createMaterialGQL, test.Object{
		"input": test.Object{
			"name":          "1828",
			"customerCode":  "613-12760",
			"projectRemark": "D53 PRL TOP",
			"points": []test.Object{
				{
					"name":       "FAI3_G7",
					"upperLimit": 5.36,
					"nominal":    5.31,
					"lowerLimit": 5.26,
				},
				{
					"name":       "FAI3_G8",
					"upperLimit": 5.36,
					"nominal":    5.31,
					"lowerLimit": 5.26,
				},
				{
					"name":       "FAI4_G1",
					"upperLimit": 4.28,
					"nominal":    4.23,
					"lowerLimit": 4.18,
				},
				{
					"name":       "FAI4_G2",
					"upperLimit": 4.28,
					"nominal":    4.23,
					"lowerLimit": 4.18,
				},
				{
					"name":       "FAI4_G3",
					"upperLimit": 4.28,
					"nominal":    4.23,
					"lowerLimit": 4.18,
				},
			},
		},
	}).GQLObject().Path("$.data.response")
	id := ret.Object().Value("id").Number().Raw()
	return uint(id)
}
