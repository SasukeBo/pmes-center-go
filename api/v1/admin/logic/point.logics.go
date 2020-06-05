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

func SavePoints(ctx context.Context, materialID int, saveItems []*model.PointCreateInput, deleteItems []int) (model.ResponseStatus, error) {
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if !user.IsAdmin {
		return model.ResponseStatusError, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	tx := orm.Begin()

	for _, id := range deleteItems {
		if err := tx.Delete(orm.Point{}, "id = ?", id).Error; err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(gc, errormap.ErrorCodeDeleteObjectError, err, "point")
		}
	}

	for _, saveItem := range saveItems {
		var point orm.Point
		if saveItem.ID != nil {
			point.Get(uint(*saveItem.ID))
		}
		point.MaterialID = uint(materialID)
		point.Name = saveItem.Name
		point.UpperLimit = saveItem.Usl
		point.LowerLimit = saveItem.Lsl
		point.Nominal = saveItem.Nominal
		if err := tx.Save(&point).Error; err != nil {
			tx.Rollback()
			return model.ResponseStatusError, errormap.SendGQLError(gc, errormap.ErrorCodeSaveObjectError, err, "point")
		}
	}

	tx.Commit()
	return model.ResponseStatusOk, nil
}

func ListMaterialPoints(ctx context.Context, materialID int, page int, limit int) (*model.PointWrap, error) {
	gc := api.GetGinContext(ctx)
	user := api.CurrentUser(gc)
	if !user.IsAdmin {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodePermissionDeny, nil)
	}

	sql := orm.Model(&orm.Point{}).Where("material_id = ?", materialID)

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeCountObjectFailed, err, "material_points")
	}

	var points []orm.Point
	if err := sql.Offset((page - 1) * limit).Limit(limit).Find(&points).Error; err != nil {
		return nil, errormap.SendGQLError(gc, errormap.ErrorCodeGetObjectFailed, err, "material_points")
	}

	var outs []*model.Point
	for _, point := range points {
		var out model.Point
		if err := copier.Copy(&out, &point); err != nil {
			continue
		}

		outs = append(outs, &out)
	}

	return &model.PointWrap{
		Points: outs,
		Total:  total,
	}, nil
}
