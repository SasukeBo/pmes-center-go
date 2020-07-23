package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
)

func MaterialVersions(ctx context.Context, id int, search *string, limit *int) ([]*model.MaterialVersion, error) {
	var versions []orm.MaterialVersion
	query := orm.Model(&orm.MaterialVersion{}).Where("material_id = ?", id).Order("created_at desc")
	if search != nil {
		query = query.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *search))
	}
	if limit != nil {
		query = query.Limit(*limit)
	}

	if err := query.Find(&versions).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
	}

	var outs []*model.MaterialVersion
	for _, v := range versions {
		var out model.MaterialVersion
		if err := copier.Copy(&out, &v); err != nil {
			log.Errorln(err)
			continue
		}
		outs = append(outs, &out)
	}

	return outs, nil
}
