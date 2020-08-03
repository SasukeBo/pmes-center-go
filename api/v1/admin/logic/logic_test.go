package logic

import (
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"testing"
	"time"
)

func TestCharASCII(t *testing.T) {
	var str = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var chars = []rune(str)
	for _, char := range chars {
		fmt.Printf("char %s ASCII is %v\n", string(char), char)
	}
}

func TestParseIndexInCodeRange(t *testing.T) {
	index, err := parseIndexInCodeRange("G", "5", "E", "1", "2", "8", "A", "C")
	fmt.Println(index, err)
}

func TestWeekDayTime(t *testing.T) {
	now := time.Now()
	pt := time.Date(now.Year(), time.January, 31*7, 0, 0, 0, 0, time.UTC)
	fmt.Println(pt)
}

func TestBarCodeDecoder_Decode(t *testing.T) {
	orm.DB.Exec("delete from bar_code_rules where 1 = 1")
	itemsMap := make(types.Map)
	items := []orm.BarCodeItem{
		{
			Label:      "厂商代码",
			Key:        "Attr1",
			IndexRange: []int{1, 3},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "年份",
			Key:        "Attr2",
			IndexRange: []int{4},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "焊接周别",
			Key:        "Attr3",
			IndexRange: []int{5, 6},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "焊接周别天数",
			Key:        "Attr4",
			IndexRange: []int{7},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "连续计数",
			Key:        "Attr5",
			IndexRange: []int{8, 11},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "工程配置码",
			Key:        "Attr6",
			IndexRange: []int{12, 15},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "图纸版本",
			Key:        "Attr7",
			IndexRange: []int{16},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "焊接线编号",
			Key:        "Line",
			IndexRange: []int{18},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "焊接治具编号",
			Key:        "Fixture",
			IndexRange: []int{19, 20},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "班别",
			Key:        "Shift",
			IndexRange: []int{21},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "主支架模号",
			Key:        "Tool1",
			IndexRange: []int{22},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "主支架生产日期",
			Key:        "Date1",
			IndexRange: []int{23, 24},
			Type:       model.BarCodeItemTypeDatetime.String(),
			DayCode:    []string{"1", "Z"},
			MonthCode:  []string{"1", "C"},
		},
		{
			Label:      "辅支架模号",
			Key:        "Tool2",
			IndexRange: []int{25},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "辅支架生产日期",
			Key:        "Date2",
			IndexRange: []int{26, 27},
			Type:       model.BarCodeItemTypeDatetime.String(),
			DayCode:    []string{"1", "Z"},
			MonthCode:  []string{"1", "C"},
		},
		{
			Label:      "原材料厂商",
			Key:        "Supplier",
			IndexRange: []int{28},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
	}
	itemsMap["items"] = items
	var rule = orm.BarCodeRule{
		CodeLength: 0,
		Name:       "TestBarCodeRule",
		Remark:     "测试解析规则",
		Items:      itemsMap,
	}
	orm.Create(&rule)
	var queryRule orm.BarCodeRule
	queryRule.Get(rule.ID)

	decoder := NewBarCodeDecoder(&queryRule)

	//          0        1         2
	//          1234567890123456789012345678
	var code = "FTTA31703E42867H1102B17K17MT"
	fmt.Printf("	开始解析：%v\n", code)
	result, statusCode := decoder.Decode(code)
	switch statusCode {
	case 1:
		log.Info("Decode successful.")
	case 2:
		log.Error("Decode failed.")
	case 3:
		log.Error("Empty code")
	}

	fmt.Println()
	fmt.Println("	解析结果：")
	for _, item := range items {
		fmt.Printf("	%s: %v\n", item.Label, result[item.Key])
	}
}
