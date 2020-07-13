package logic

import (
	"context"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
)

func LoadFile(ctx context.Context, id uint) *model.File {
	var file orm.File
	if err := file.Get(id); err != nil {
		return nil
	}

	var out model.File
	if err := copier.Copy(&out, &file); err != nil {
		return nil
	}

	return &out
}
