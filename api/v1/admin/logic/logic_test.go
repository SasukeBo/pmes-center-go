package logic

import (
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/tealeg/xlsx/v3"
	"io/ioutil"
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
	fmt.Printf("第1周 周日是：%v\n", parseTimeFromWeekday(1, 1-1))
	fmt.Printf("第2周 周五是：%v\n", parseTimeFromWeekday(2, 6-1))
	fmt.Printf("第30周 周一是：%v\n", parseTimeFromWeekday(30, 2-1))
	fmt.Printf("第31周 周六是：%v\n", parseTimeFromWeekday(31, 7-1))
	fmt.Printf("第32周 周一是：%v\n", parseTimeFromWeekday(32, 2-1))
}

func TestBarCodeDecoder_Decode(t *testing.T) {
	orm.DB.Exec("delete from bar_code_rules where 1 = 1")
	itemsMap := make(types.Map)
	items := []orm.BarCodeItem{
		{
			Label:       "厂商代码",
			Key:         "Attr1",
			IndexRange:  []int{1, 3},
			CategorySet: []string{"FTT"},
			Type:        model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "年份",
			Key:        "Attr2",
			IndexRange: []int{4},
			Type:       model.BarCodeItemTypeCategory.String(),
		},
		{
			Label:      "焊接日期",
			Key:        "Attr3",
			IndexRange: []int{5, 7},
			Type:       model.BarCodeItemTypeWeekday.String(),
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
			Label:           "主支架生产日期",
			Key:             "Date1",
			IndexRange:      []int{23, 24},
			Type:            model.BarCodeItemTypeDatetime.String(),
			DayCode:         []string{"1", "Z"},
			MonthCode:       []string{"1", "C"},
			MonthCodeReject: []string{"6", "B"},
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
		CodeLength: 28,
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
	var code = "FTTA05603E42867H1102B17K17MT"
	fmt.Printf("	开始解析：%v\n", code)
	result, statusCode := decoder.Decode(code)
	switch statusCode {
	case 1:
		log.Info("正确成功.")
	case 2:
		log.Error("识别码不符合编码规则")
	case 3:
		log.Error("识别码读取失败，为空字符串")
	case 4:
		log.Error("识别码长度不正确")
	default:
		log.Error("Got Error %v", statusCode)
	}

	fmt.Println()
	fmt.Println("	解析结果：")
	for _, item := range items {
		fmt.Printf("	%s: %v\n", item.Label, result[item.Key])
	}
}

func TestAssembleDataIntoFile(t *testing.T) {
	var name = "abc"
	var points []orm.Point
	for i := 0; i < 10; i++ {
		var point = orm.Point{
			Name:       fmt.Sprintf("point_%d", i),
			Index:      parseIndexFromColumnCode("J") + i,
			UpperLimit: float64(i) * 0.54,
			LowerLimit: float64(i) * 0.87,
			Nominal:    float64(i) * 0.67,
		}
		points = append(points, point)
	}
	var products []orm.Product
	for i := 0; i < 10; i++ {
		var p = orm.Product{
			Qualified: i%9 == 0,
			BarCode:   "ABC",
			CreatedAt: time.Now(),
		}
		var pv = make(types.Map)
		for i, point := range points {
			pv[point.Name] = i
		}
		p.PointValues = pv
		products = append(products, p)
	}

	assembleDataIntoFile(name, points, products)
}

func TestReadXlsxStyle(t *testing.T) {
	content, err := ioutil.ReadFile("/Users/sasukebo/workspace/pmes_data_cache/cache/test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	file, err := xlsx.OpenBinary(content)
	if err != nil {
		t.Fatal(err)
	}

	sheet := file.Sheets[0]
	row, _ := sheet.Row(0)
	style := row.GetCell(0).GetStyle()
	fmt.Println(style)
	rowData, _ := sheet.Row(11)
	style = rowData.GetCell(10).GetStyle()
	fmt.Println(style)
}
