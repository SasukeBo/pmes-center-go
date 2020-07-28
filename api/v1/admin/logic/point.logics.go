package logic

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/SasukeBo/pmes-data-center/api"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
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
		var name string
		name = dimRow.Cells[i].String()
		if name == "" { // 过滤空单元格
			continue
		}

		var usl, lsl, nominal float64
		if usl, err = uslRow.Cells[i].Float(); err != nil {
			usl = 0
		}
		if nominal, err = nominalRow.Cells[i].Float(); err != nil {
			nominal = 0
		}
		if lsl, err = lslRow.Cells[i].Float(); err != nil {
			lsl = 0
		}
		out := model.Point{
			Name:       name,
			UpperLimit: usl,
			LowerLimit: lsl,
			Nominal:    nominal,
			Index:      parseColumnCodeFromIndex(i + 1),
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
		point.Index = parseIndexFromColumnCode(saveItem.Index)
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

func ListMaterialPoints(ctx context.Context, materialVersionID int) ([]*model.Point, error) {
	user := api.CurrentUser(ctx)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodePermissionDeny, nil)
	}

	var points []orm.Point
	query := orm.Model(&orm.Point{}).Where("material_version_id = ?", materialVersionID)
	query = query.Order("points.index ASC")
	if err := query.Find(&points).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_points")
	}

	var outs []*model.Point
	for _, point := range points {
		var out model.Point
		if err := copier.Copy(&out, &point); err != nil {
			continue
		}
		out.Index = parseColumnCodeFromIndex(point.Index)
		outs = append(outs, &out)
	}

	return outs, nil
}
