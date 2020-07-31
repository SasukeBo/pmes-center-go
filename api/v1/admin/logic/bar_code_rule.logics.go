package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
	"strconv"
)

const (
	itemsMapKey = "items"
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
					var outItem model.BarCodeItem
					outItem.Label = fmt.Sprint(item["label"])
					outItem.Type = model.BarCodeItemType(fmt.Sprint(item["type"]))
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
