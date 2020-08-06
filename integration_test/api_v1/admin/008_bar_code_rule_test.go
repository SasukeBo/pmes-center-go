package admin

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	test "github.com/SasukeBo/pmes-data-center/integration_test"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"testing"
)

const (
	saveBarCodeRuleGQL = `
	mutation($input: BarCodeRuleInput!) {
	  response: saveBarCodeRule(input: $input)
	}`

	listBarCodeRulesGQL = `
	query($search: String, $limit: Int!, $page: Int!) {
	  response: listBarCodeRules(search: $search, limit: $limit, page: $page) {
	    total
	    rules {
	      id
	      codeLength
	      name
	      remark
	      user {
	        id
	        account
	      }
	      items {
	        label
	        key
	        indexRange
	        type
	        dayCode
	        monthCode
	      }
	      createdAt
	    }
	  }
	}
	`

	getBarCodeRuleGQL = `
	query($id: Int!) {
	  response: getBarCodeRule(id: $id) {
	    id
	    codeLength
	    name
	    remark
	    user {
	      id
	      account
	    }
	    items {
	      label
	      key
	      indexRange
	      type
	      dayCode
	      monthCode
		  categorySet
	    }
	    createdAt
	  }
	}
	`
)

func TestBarCodeRule(t *testing.T) {
	tester := test.NewTester(t)
	test.Login(test.AdminAccount, test.AdminPasswd, true)
	ids := generateBarCodeRules()

	t.Run("SAVE_BAR_CODE_RULE_USE_RESERVED_KEY", func(t *testing.T) {
		tester.API1Admin(saveBarCodeRuleGQL, test.Object{
			"input": test.Object{
				"name":       "process save rule failed",
				"remark":     "测试编码规则",
				"codeLength": 28,
				"items": []test.Object{
					{
						"label":      "冲压班别",
						"key":        "Shift",
						"type":       "Category",
						"indexRange": []int{1},
					},
				},
			},
		}).GQLObject().Path("$.data").Null()
	})

	t.Run("SAVE_BAR_CODE_RULE", func(t *testing.T) {
		tester.API1Admin(saveBarCodeRuleGQL, test.Object{
			"input": test.Object{
				"name":       "process rule",
				"remark":     "测试编码规则",
				"codeLength": 28,
				"items": []test.Object{
					{
						"label":       "治具号",
						"key":         "Fixture",
						"type":        "Category",
						"indexRange":  []int{1},
						"categorySet": []string{"A", "B"},
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
	})

	t.Run("LIST_BAR_CODE_RULES", func(t *testing.T) {
		ret := tester.API1Admin(listBarCodeRulesGQL, test.Object{
			"search": "fake",
			"limit":  4,
			"page":   1,
		}).GQLObject().Path("$.data.response").Object()
		ret.Value("total").Equal(5)
		ret.Value("rules").Array().Length().Equal(4)
	})

	t.Run("GET_BAR_CODE_RULE", func(t *testing.T) {
		ret := tester.API1Admin(getBarCodeRuleGQL, test.Object{"id": ids[0]}).GQLObject().Path("$.data.response")
		ret.Object().Value("id").Equal(ids[0])
	})
}

func generateBarCodeRules() []uint {
	var ids []uint
	for i := 0; i < 5; i++ {
		itemsMap := make(types.Map)
		items := []orm.BarCodeItem{
			{
				Label:       fmt.Sprintf("label_%v_1", i),
				Key:         fmt.Sprintf("key_%v_1", i),
				IndexRange:  []int{i, i + 1},
				Type:        model.BarCodeItemTypeCategory.String(),
				CategorySet: []string{"A", "B"},
			},
			{
				Label:      fmt.Sprintf("label_%v_2", i),
				Key:        fmt.Sprintf("key_%v_2", i),
				IndexRange: []int{i + 1, i + 2},
				Type:       model.BarCodeItemTypeDatetime.String(),
				DayCode:    []string{"1", "V"},
				MonthCode:  []string{"1", "A"},
			},
		}
		itemsMap["items"] = items
		var rule = orm.BarCodeRule{
			CodeLength: 0,
			Name:       fmt.Sprintf("fake_rule_name_%v", i),
			Remark:     fmt.Sprintf("fake_rule_remark_%v", i),
			UserID:     test.Data.Admin.ID,
			Items:      itemsMap,
		}
		orm.Create(&rule)
		ids = append(ids, rule.ID)
	}

	return ids
}
