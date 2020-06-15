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
	"strings"
)

func ParseImportPoints(ctx context.Context, file graphql.Upload) ([]*model.Point, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	if filepath.Ext(file.Filename) != ".xlsx" {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeFileExtensionError, nil, ".xlsx")
	}

	content, err := ioutil.ReadAll(file.File)
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeFileHandleError, err)
	}

	xFile, err := xlsx.OpenBinary(content)
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeFileHandleError, err)
	}

	sheet := xFile.Sheets[0]
	dimRow := sheet.Row(0)
	uslRow := sheet.Row(1)
	nominalRow := sheet.Row(2)
	lslRow := sheet.Row(3)

	var outs []*model.Point
	for i := 1; i < len(dimRow.Cells); i++ {
		name := dimRow.Cells[i].String()
		usl, _ := uslRow.Cells[i].Float()
		nominal, _ := nominalRow.Cells[i].Float()
		lsl, _ := lslRow.Cells[i].Float()
		out := model.Point{
			Name:       name,
			UpperLimit: usl,
			LowerLimit: lsl,
			Nominal:    nominal,
		}
		outs = append(outs, &out)
	}

	return outs, nil
}

func SavePoints(ctx context.Context, materialID int, saveItems []*model.PointCreateInput, deleteItems []int) (model.ResponseStatus, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()

	for _, id := range deleteItems {
		if err := tx.Delete(orm.Point{}, "id = ?", id).Error; err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeDeleteObjectError, err, "point")
		}
	}

	for _, saveItem := range saveItems {
		var point orm.Point

		if saveItem.ID != nil && *saveItem.ID != 0 {
			point.Get(uint(*saveItem.ID))
		}

		point.MaterialID = uint(materialID)
		point.Name = saveItem.Name
		point.UpperLimit = saveItem.UpperLimit
		point.LowerLimit = saveItem.LowerLimit
		point.Nominal = saveItem.Nominal
		if err := tx.Save(&point).Error; err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "Error 1062") {
				return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodePointAlreadyExists, err)
			}
			return model.ResponseStatusError, errormap.SendGQLError(ctx, errormap.ErrorCodeSaveObjectError, err, "point")
		}
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func ListMaterialPoints(ctx context.Context, materialID int) ([]*model.Point, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var points []orm.Point
	if err := orm.Model(&orm.Point{}).Where("material_id = ?", materialID).Find(&points).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_points")
	}

	var outs []*model.Point
	for _, point := range points {
		var out model.Point
		if err := copier.Copy(&out, &point); err != nil {
			continue
		}

		outs = append(outs, &out)
	}

	return outs, nil
}
