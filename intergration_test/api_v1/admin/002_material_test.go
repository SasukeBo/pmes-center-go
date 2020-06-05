package admin

import (
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/intergration_test"
	"github.com/SasukeBo/ftpviewer/orm"
	"testing"
	"time"
)

// NOTE: this test need your ftp service working
// - docker-compose up ftp
// - then create a directory named 1828
// - put a data file into this directory
func TestMaterial(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, true)

	// Create material and load data
	t.Run("CREATE_MATERIAL_AND_LOAD_DATA", func(t *testing.T) {
		go ftpclient.FTPWorker()
		tester.API1Admin(createMaterialGQL, test.Object{
			"input": test.Object{
				"name":          "1828",
				"customerCode":  "613-12760",
				"projectRemark": "D53 PRL TOP",
				"points": []test.Object{
					{
						"name":    "FAI3_G7",
						"usl":     5.36,
						"nominal": 5.31,
						"lsl":     5.26,
					},
					{
						"name":    "FAI3_G8",
						"usl":     5.36,
						"nominal": 5.31,
						"lsl":     5.26,
					},
					{
						"name":    "FAI4_G1",
						"usl":     4.28,
						"nominal": 4.23,
						"lsl":     4.18,
					},
					{
						"name":    "FAI4_G2",
						"usl":     4.28,
						"nominal": 4.23,
						"lsl":     4.18,
					},
					{
						"name":    "FAI4_G3",
						"usl":     4.28,
						"nominal": 4.23,
						"lsl":     4.18,
					},
				},
			},
		}).GQLObject().Path("$.data.response")
		// wait for worker
		<-time.After(4 * time.Second)
	})

	// query materials with pattern
	t.Run("QUERY_MATERIALS_WITH_PATTERN", func(t *testing.T) {
		orm.Create(&orm.Material{
			Name:          "test_material_1",
			CustomerCode:  "test_customer_code_1",
			ProjectRemark: "apple",
		})
		orm.Create(&orm.Material{
			Name:          "test_material_2",
			CustomerCode:  "test_customer_code_2",
			ProjectRemark: "apply",
		})
		orm.Create(&orm.Material{
			Name:          "test_material_3",
			CustomerCode:  "test_customer_code_3",
			ProjectRemark: "application",
		})
		orm.Create(&orm.Material{
			Name:          "test_material_4",
			CustomerCode:  "test_customer_code_4",
			ProjectRemark: "hello",
		})

		ret := tester.API1Admin(listMaterialGQL, test.Object{
			"pattern": "app",
			"page":    1,
			"limit":   2,
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("total").Equal(3)
		ret.Value("materials").Array().Length().Equal(2)
	})

	// delete material
	t.Run("DELETE_MATERIAL", func(t *testing.T) {
		material := orm.Material{
			Name:          "test_material_1",
			CustomerCode:  "test_customer_code_1",
			ProjectRemark: "apple",
		}
		orm.Create(&material)
		tester.API1Admin(deleteMaterialGQL, test.Object{"id": material.ID}).GQLObject().Path("$.data.response").Equal("OK")
	})

	// update material
	t.Run("UPDATE_MATERIAL", func(t *testing.T) {
		material := orm.Material{
			Name:          "test_material_1",
			CustomerCode:  "test_customer_code_1",
			ProjectRemark: "apple",
		}
		orm.Create(&material)
		ret := tester.API1Admin(updateMaterialGQL, test.Object{
			"input": test.Object{
				"id":            material.ID,
				"customerCode":  "changed customer code",
				"projectRemark": "changed project remark",
			},
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("id").Equal(material.ID)
		ret.Value("customerCode").Equal("changed customer code")
		ret.Value("createdAt").NotNull()
	})
}
