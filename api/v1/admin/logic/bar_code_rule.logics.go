package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/jinzhu/copier"
	"strconv"
	"time"
)

const (
	itemsMapKey = "items"
)

var (
	reservedCategory = []string{"Date", "Device", "Shift", "Attribute"}
)

func SaveBarCodeRule(ctx context.Context, input model.BarCodeRuleInput) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var rule orm.BarCodeRule
	rule.UserID = user.ID
	if input.ID != nil {
		if err := rule.Get(uint(*input.ID)); err != nil {
			return model.ResponseStatusError, errormap.SendGQLError(ctx, err.GetCode(), err, "bar_code_rule")
		}
	}

	rule.Name = input.Name
	rule.Remark = input.Remark
	rule.CodeLength = input.CodeLength

	rule.Items = make(types.Map)
	var items []orm.BarCodeItem

	for _, itemInput := range input.Items {
		if util.Includes(reservedCategory, itemInput.Key) {
			return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeBarCodeReservedKey, nil, itemInput.Key)
		}

		var item orm.BarCodeItem
		if err := copier.Copy(&item, &itemInput); err != nil {
			continue
		}

		items = append(items, item)
	}
	rule.Items[itemsMapKey] = items
	if err := orm.Save(&rule).Error; err != nil {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "bar_code_rule")
	}

	return model.ResponseStatusOk, nil
}

func ListBarCodeRules(ctx context.Context, search *string, limit int, page int) (*model.BarCodeRuleWrap, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var query = orm.Model(&orm.BarCodeRule{})
	if search != nil {
		var pattern = fmt.Sprintf("%%%s%%", *search)
		query = query.Where("name like ? OR remark like ?", pattern, pattern)
	}

	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "bar_code_rule")
	}

	var rules []orm.BarCodeRule
	if err := query.Limit(limit).Offset((page - 1) * limit).Find(&rules).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "bar_code_rule")
	}

	var outs []*model.BarCodeRule
	for _, r := range rules {
		out := convertBarCodeRule(&r)
		outs = append(outs, &out)
	}

	return &model.BarCodeRuleWrap{
		Total: total,
		Rules: outs,
	}, nil
}

func convertBarCodeRule(rule *orm.BarCodeRule) model.BarCodeRule {
	var out model.BarCodeRule
	if err := copier.Copy(&out, rule); err != nil {
		log.Error("copy BarCodeRule failed: %v\n", err)
		return out
	}

	var items []*model.BarCodeItem
	if v, ok := rule.Items["items"]; ok {
		if inputItems, ok := v.([]interface{}); ok {
			for _, inputItem := range inputItems {
				if item, ok := inputItem.(map[string]interface{}); ok {
					outDBItem := DecodeBarCodeItemFromDBToStruct(item)
					var outItem model.BarCodeItem
					if err := copier.Copy(&outItem, &outDBItem); err != nil {
						log.Errorln(err)
						continue
					}
					outItem.Type = model.BarCodeItemType(outDBItem.Type)
					items = append(items, &outItem)
				}
			}
		}
	}

	out.Items = items
	return out
}

func GetBarCodeRule(ctx context.Context, id int) (*model.BarCodeRule, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var rule orm.BarCodeRule
	if err := rule.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "bar_code_rule")
	}

	out := convertBarCodeRule(&rule)
	return &out, nil
}

func LoadBarCodeRule(ctx context.Context, id uint) *model.BarCodeRule {
	var rule orm.BarCodeRule
	if err := rule.Get(id); err != nil {
		log.Errorln(err)
		return nil
	}

	out := convertBarCodeRule(&rule)
	return &out
}

func DecodeBarCodeItemFromDBToStruct(item map[string]interface{}) orm.BarCodeItem {
	var outItem orm.BarCodeItem
	outItem.Label = fmt.Sprint(item["label"])
	outItem.Type = fmt.Sprint(item["type"])
	outItem.Key = fmt.Sprint(item["key"])
	if codes, ok := item["day_code"].([]interface{}); ok {
		var dayCode []string
		for _, code := range codes {
			dayCode = append(dayCode, fmt.Sprint(code))
		}
		outItem.DayCode = dayCode
	}
	if codes, ok := item["month_code"].([]interface{}); ok {
		var monthCode []string
		for _, code := range codes {
			monthCode = append(monthCode, fmt.Sprint(code))
		}
		outItem.MonthCode = monthCode
	}
	if codes, ok := item["index_range"].([]interface{}); ok {
		var indexRange []int
		for _, c := range codes {
			code, err := strconv.Atoi(fmt.Sprint(c))
			if err != nil {
				code = 0
			}
			indexRange = append(indexRange, code)
		}
		outItem.IndexRange = indexRange
	}

	return outItem
}

type BarCodeDecoder struct {
	Rules       []orm.BarCodeItem
	BarCodeRule *orm.BarCodeRule
}

func NewBarCodeDecoder(rule *orm.BarCodeRule) BarCodeDecoder {
	var decoder BarCodeDecoder
	itemsMapValue, ok := rule.Items["items"]
	if !ok {
		return decoder
	}
	items, ok := itemsMapValue.([]interface{})
	if !ok {
		return decoder
	}
	var rules []orm.BarCodeItem
	for _, v := range items {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		out := DecodeBarCodeItemFromDBToStruct(item)
		rules = append(rules, out)
	}

	decoder.Rules = rules
	decoder.BarCodeRule = rule
	return decoder
}

// Decode 解析识别码，返回解析结果对象 及 状态码
// 状态码：
// - 1 正确识别
// - 2 识别码不符合编码规则
// - 3 识别码读取失败，为空字符串
// - 4 识别码长度不正确
func (bdc *BarCodeDecoder) Decode(code string) (out types.Map, statusCode int) {
	out = make(types.Map)
	if code == "" {
		statusCode = orm.BarCodeStatusEmpty
		return
	}
	if len(code) != bdc.BarCodeRule.CodeLength {
		statusCode = orm.BarCodeStatusTooShort
		return
	}

	for _, rule := range bdc.Rules {
		var begin, end int
		if len(rule.IndexRange) > 0 {
			begin = rule.IndexRange[0]
		}
		if len(rule.IndexRange) > 1 {
			end = rule.IndexRange[1]
		}
		var childStr string
		if end != 0 {
			childStr = code[begin-1 : end]
		} else {
			childStr = string(code[begin-1])
		}

		switch rule.Type {
		case "Category":
			out[rule.Key] = childStr
		case "Datetime":
			timeCode := childStr
			var t *time.Time
			var err error

			if len(timeCode) > 1 {
				t, err = parseCodeDatetime(timeCode[:1], timeCode[1:2], rule)
			} else if len(timeCode) > 0 {
				t, err = parseCodeDatetime("", timeCode, rule)
			}

			if err != nil {
				statusCode = orm.BarCodeStatusIllegal
				return
			}

			if t == nil {
				out[rule.Key] = time.Now()
			} else {
				out[rule.Key] = *t
			}
		}
	}

	statusCode = orm.BarCodeStatusSuccess
	return
}

func parseCodeDatetime(monthCode, dayCode string, rule orm.BarCodeItem) (*time.Time, error) {
	var month, day int
	var err error

	if monthCode != "" && len(rule.MonthCode) > 1 {
		month, err = parseIndexInCodeRange(monthCode, rule.MonthCode[0], rule.MonthCode[1], rule.MonthCode[2:]...)
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		if month > 12 {
			err = errors.New(fmt.Sprintf("month out range of 1 - 12, got %v", month))
			log.Errorln(err)
			return nil, err
		}
	}

	if dayCode != "" && len(rule.DayCode) > 1 {
		day, err = parseIndexInCodeRange(dayCode, rule.DayCode[0], rule.DayCode[1], rule.DayCode[2:]...)
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		if day > 31 {
			err = errors.New(fmt.Sprintf("month out range of 1 - 31, got %v", day))
			log.Errorln(err)
			return nil, err
		}
	}

	now := time.Now()
	if month == 0 {
		month = int(now.Month())
	}
	if day == 0 {
		day = now.Day()
	}

	t := time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t, nil
}

func parseIndexInCodeRange(code, begin, end string, rejects ...string) (int, error) {
	ascii := code[0]
	var distance uint8
	if ascii >= begin[0] && ascii <= end[0] {
		distance = ascii - begin[0]
		for _, r := range rejects {
			if r[0] == ascii {
				return 0, errors.New("cannot parse rejected code")
			}
			if r[0] < ascii && r[0] >= begin[0] {
				distance--
			}
		}
		if ascii >= uint8('A') {
			distance = distance - 7
		}
	} else {
		return 0, errors.New("code is out range")
	}

	return int(distance + 1), nil
}

func parseTimeFromWeekday(week, day int) *time.Time {
	now := time.Now()
	t := time.Date(now.Year(), time.January, 7*(week-1), 0, 0, 0, 0, time.UTC)
	weekDay := t.Weekday()
	distance := int(day) - int(weekDay)
	nt := t.AddDate(0, 0, distance)
	return &nt
}
