package admin

import (
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"github.com/SasukeBo/pmes-data-center/orm"
	"testing"
)

const (
	updateDecodeTemplateGQL = `
	mutation($input: DecodeTemplateInput!) {
	  response: updateDecodeTemplate(input: $input)
	}
	`
	decodeTemplateWithVersionIDGQL = `
	query($id: Int!) {
	  response: decodeTemplateWithVersionID(id: $id) {
	    id
	    barCodeRule {
	      id
	      codeLength
	      name
	      remark
	      items {
	        label
	        key
	        indexRange
	        type
	        dayCode
	        monthCode
	      }
	    }
	  }
	}
	`
)

func TestDecodeTemplate(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, false)
	ruleIDs := generateBarCodeRules()

	// Test save decode_template
	t.Run("TEST_SAVE_DECODE_TEMPLATE", func(t *testing.T) {
		template := orm.DecodeTemplate{
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			DataRowIndex:         15,
			CreatedAtColumnIndex: 2,
		}
		orm.Create(&template)
		tester.API1Admin(updateDecodeTemplateGQL, test.Object{
			"input": test.Object{
				"id":                   template.ID,
				"dataRowIndex":         15,
				"createdAtColumnIndex": "C",
				"barCodeRuleID":        ruleIDs[0],
				"barCodeIndex":         "B",
				"pointColumns":         []test.Object{},
			},
		}).GQLObject().Path("$.data.response").Equal("OK")
	})

	t.Run("TEST_decodeTemplateWithVersionID", func(t *testing.T) {
		version := orm.MaterialVersion{
			Version:     "test version",
			Description: "test version description",
			MaterialID:  test.Data.Material.ID,
			Active:      true,
		}
		orm.Create(&version)
		template := orm.DecodeTemplate{
			MaterialID:           test.Data.Material.ID,
			UserID:               test.Data.User.ID,
			MaterialVersionID:    version.ID,
			BarCodeRuleID:        ruleIDs[0],
			DataRowIndex:         15,
			CreatedAtColumnIndex: 2,
		}
		orm.Create(&template)
		tester.API1Admin(decodeTemplateWithVersionIDGQL, test.Object{"id": version.ID}).GQLObject()
	})
}
