package logic

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/SasukeBo/ftpviewer/api"
	"github.com/SasukeBo/ftpviewer/api/v1/admin/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"path/filepath"
)

func ImportPoints(ctx context.Context, file graphql.Upload, materialID int) ([]*model.Point, error) {
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	if filepath.Ext(file.Filename) != ".xlsx" {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeFileExtensionError, nil, ".xlsx")
	}

	content, err := ioutil.ReadAll(file.File)
	if err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeFileHandleError, err)
	}

	xFile, err := xlsx.OpenBinary(content)
	if err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeFileHandleError, err)
	}

	sheet := xFile.Sheets[0]
	dimRow := sheet.Row(0)
	uslRow := sheet.Row(1)
	nominalRow := sheet.Row(2)
	lslRow := sheet.Row(3)

	var outs []*model.Point
	tx := orm.Begin()
	for i := 1; i < len(dimRow.Cells); i++ {
		name := dimRow.Cells[i].String()
		usl, _ := uslRow.Cells[i].Float()
		nominal, _ := nominalRow.Cells[i].Float()
		lsl, _ := lslRow.Cells[i].Float()
		point := orm.Point{
			Name:       name,
			MaterialID: uint(materialID),
			UpperLimit: usl,
			LowerLimit: lsl,
			Nominal:    nominal,
		}
		if err := orm.Create(&point).Error; err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCreateObjectError, err, "point")
		}
		var out model.Point
		if err := copier.Copy(&out, &point); err != nil {
			tx.Rollback()
			return nil, errormap.SendGQLError(gc, errormap.ErrorCodeTransferObjectError, err, "point")
		}
		outs = append(outs, &out)
	}
	tx.Commit()

	return outs, nil
}
