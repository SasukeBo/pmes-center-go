package logic

import (
	"context"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
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
