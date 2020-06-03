package test

import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestPointsImport(t *testing.T) {
	tester := newTester(t)
	material := &orm.Material{
		Name:          "test_material",
		CustomerCode:  "test_material_customer_code",
		ProjectRemark: "test_material_project_remark",
	}

	login(testAdminAccount, testAdminPasswd, false)
	orm.Create(material)

	path, _ := os.Getwd()
	file, err := os.Open(filepath.Join(path, "../priv/import_points_template.xlsx"))
	if err != nil {
		t.Fatalf("open file failed: %v\n", err)
	}

	formData := object{
		"operations": fmt.Sprintf("{\"query\": \"%s\", \"variables\": {\"materialID\": %v, \"file\": null }}", pointImportGQL, material.ID),
		"map":        `{"template": ["variables.file"]}`,
	}
	ret := tester.Upload("/api/v1").WithMultipart().WithForm(formData).WithFile("template", "import_points_template.xlsx", file).Expect().Status(http.StatusOK)
	ret1 := ret.JSON().Object().Path("$.data.response").Array()
	ret1.Length().Equal(19)
	ret1.First().Object().Value("name").Equal("FAI3_G7")
	ret1.Last().Object().Value("name").Equal("Profile")
}
