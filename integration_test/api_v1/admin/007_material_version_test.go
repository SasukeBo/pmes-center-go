package admin

import (
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"testing"
)

var decodeTemplateInput = test.Object{
	"materialID":           test.Data.Material.ID,
	"dataRowIndex":         15,
	"createdAtColumnIndex": "B",
}

func TestMaterialVersion(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, true)

	t.Run("CREATE_MATERIAL_VERSION", func(t *testing.T) {
		tester.API1Admin(GqlCreateMaterialVersion, test.Object{
			"input": test.Object{
				"materialID":  test.Data.Material.ID,
				"version":     "process version",
				"description": "process description",
				"active":      true,
				"points":      inputPoints,
				"template":    decodeTemplateInput,
			},
		}).GQLObject().Path("$.data.response").String().Equal("OK")
	})
}

var GqlCreateMaterialVersion = `
mutation($input: MaterialVersionInput!) {
	response: createMaterialVersion(input: $input)
}
`
