package admin

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/integration_test"
	"github.com/SasukeBo/pmes-data-center/orm"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestPointsImport(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)

	// import points from xlsx file
	t.Run("IMPORT_POINTS_FROM_XLSX_FILE", func(t *testing.T) {
		path, _ := os.Getwd()
		file, err := os.Open(filepath.Join(path, "./import_points_template.xlsx"))
		if err != nil {
			t.Fatalf("open file failed: %v\n", err)
		}

		formData := test.Object{
			"operations": fmt.Sprintf("{\"query\": \"%s\", \"variables\": {\"file\": null }}", pointImportParseGQL),
			"map":        `{"template": ["variables.file"]}`,
		}
		ret := tester.Upload("/api/v1/admin").WithMultipart().WithForm(formData).WithFile("template", "import_points_template.xlsx", file).Expect().Status(http.StatusOK)
		ret1 := ret.JSON().Object().Path("$.data.response").Array()
		ret1.Length().Equal(19)
		ret1.First().Object().Value("name").Equal("FAI3_G7")
		ret1.First().Object().Value("id").Equal(0)
		ret1.Last().Object().Value("name").Equal("Profile")
	})

	// process save points including delete point
	t.Run("TEST_SAVE_POINTS_INCLUDING_DELETE_POINT", func(t *testing.T) {
		point1 := orm.Point{
			Name:       "test_point_1",
			MaterialID: test.Data.Material.ID,
			UpperLimit: 10,
			LowerLimit: 1,
			Nominal:    1.1,
		}
		orm.Create(&point1)
		point2 := orm.Point{
			Name:       "test_point_2",
			MaterialID: test.Data.Material.ID,
			UpperLimit: 22,
			LowerLimit: 2,
			Nominal:    2.2,
		}
		orm.Create(&point2)

		tester.API1Admin(savePointsGQL, test.Object{
			"materialID": test.Data.Material.ID,
			"saveItems": []test.Object{
				{
					"id":      point1.ID,
					"name":    "test_point_1_name_changed",
					"usl":     11,
					"nominal": 1.11,
					"lsl":     1,
				},
				{
					"name":    "test_point_3",
					"usl":     33,
					"nominal": 3.3,
					"lsl":     3,
				},
			},
			"deleteItems": []uint{point2.ID},
		}).GQLObject().Path("$.data.response").Equal("OK")
	})

	// process get list of points with materialID
	t.Run("TEST_GET_LIST_OF_POINTS_WITH_MATERIAL_ID", func(t *testing.T) {
		var material = orm.Material{
			Name:          "TEST_GET_LIST_OF_POINTS_WITH_MATERIAL_ID",
			CustomerCode:  "TEST_GET_LIST_OF_POINTS_WITH_MATERIAL_ID",
			ProjectRemark: "TEST_GET_LIST_OF_POINTS_WITH_MATERIAL_ID",
		}
		orm.Create(&material)
		orm.Create(&orm.Point{
			Name:       "test_point_1",
			MaterialID: material.ID,
			UpperLimit: 10,
			LowerLimit: 1,
			Nominal:    1.1,
		})
		orm.Create(&orm.Point{
			Name:       "test_point_2",
			MaterialID: material.ID,
			UpperLimit: 22,
			LowerLimit: 2,
			Nominal:    2.2,
		})

		ret := tester.API1Admin(listMaterialPointGQL, test.Object{"materialID": material.ID}).GQLObject().Path("$.data.response").Array()
		ret.Length().Equal(2)
		ret.First().Object().Value("name").Equal("test_point_1")
	})
}
