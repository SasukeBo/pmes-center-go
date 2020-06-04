package logic

import (
	"context"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
)

func LoadDecodeTemplate(ctx context.Context, templateID uint) (*model.DecodeTemplate, error) {
	var template orm.DecodeTemplate
	if err := template.Get(templateID); err != nil {
		return nil, err
	}
	var out model.DecodeTemplate
	if err := copier.Copy(&out, &template); err != nil {
		return nil, err
	}

	return &out, nil
}
