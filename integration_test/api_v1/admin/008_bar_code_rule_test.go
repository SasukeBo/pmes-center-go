package admin

import (
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"testing"
)

func TestBarCodeRule(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, true)

	tester.API1Admin(saveBarCodeRuleGQL, test.Object{
		"input": test.Object{
			"name":       "test rule",
			"remark":     "测试编码规则",
			"codeLength": 28,
			"items": []test.Object{
				{
					"label":      "治具号",
					"key":        "Fixture",
					"type":       "Category",
					"indexRange": []int{1},
				},
				{
					"label":      "冲压日期",
					"key":        "ProduceDate",
					"type":       "Datetime",
					"indexRange": []int{2, 3},
					"dayCode":    []string{"1", "Y", "B", "I", "O"},
					"monthCode":  []string{"1", "C"},
				},
			},
		},
	}).GQLObject().Path("$.data.response").Equal("OK")
}

const saveBarCodeRuleGQL = `
mutation($input: BarCodeRuleInput!) {
  response: saveBarCodeRule(input: $input)
}`
