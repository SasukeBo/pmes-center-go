package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
)

func Materials(ctx context.Context, search *string, page int, limit int) (*model.MaterialsWrap, error) {
	sql := orm.Model(&orm.Material{})

	if search != nil {
		var pattern = fmt.Sprintf("%%%s%%", *search)
		sql = sql.Where("name LIKE ? OR customer_code LIKE ? OR project_remark LIKE ?", pattern, pattern, pattern)
	}

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "material")
	}

	var materials []orm.Material
	if err := sql.Offset((page - 1) * limit).Limit(limit).Find(&materials).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material")
	}

	var outs []*model.Material
	for _, m := range materials {
		var out model.Material
		if err := copier.Copy(&out, &m); err != nil {
			log.Error("[logic.Materials] copy material(%s) to out failed: %v", m.Name, err)
			continue
		}

		ok, ng := countProductQualifiedForMaterial(m.ID)
		out.Ok = ok
		out.Ng = ng
		outs = append(outs, &out)
	}

	return &model.MaterialsWrap{
		Total:     total,
		Materials: outs,
	}, nil
}

func Material(ctx context.Context, id int) (*model.Material, error) {
	var material orm.Material
	if err := material.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}

	ok, ng := countProductQualifiedForMaterial(material.ID)
	out.Ok = ok
	out.Ng = ng

	return &out, nil
}

type qualifiedResult struct {
	Qualified bool
	Total     int64
}

func countProductQualifiedForMaterial(id uint) (int, int) {
	sql := orm.Model(&orm.Product{}).Where("material_id = ?", id)
	sql = sql.Select("qualified, COUNT(id) as total")
	sql = sql.Group("qualified")
	rows, err := sql.Rows()
	if err != nil {
		_ = rows.Close()
		log.Error("[logic.countProductQualifiedForMaterial] Rows() failed: %v", err)
		return 0, 0
	}

	var ng, ok int
	for rows.Next() {
		var result qualifiedResult
		if err := rows.Scan(&result.Qualified, &result.Total); err != nil {
			log.Error("[logic.countProductQualifiedForMaterial] Scan() failed: %v", err)
			_ = rows.Close()
			return 0, 0
		}
		if result.Qualified {
			ok = int(result.Total)
		} else {
			ng = int(result.Total)
		}
	}

	_ = rows.Close()
	return ok, ng
}

func AnalyzeMaterial(ctx context.Context, analyzeInput model.AnalyzeMaterialInput) (*model.MaterialAnalysisResult, error) {
	sql := orm.Model(&orm.Product{})
	// axis
	sql = sql.Select("? as axis", analyzeInput.XAxis).Group(analyzeInput.XAxis)
	// group by
	if analyzeInput.GroupBy != nil {
		sql = sql.Select("? as group_by", *analyzeInput.GroupBy).Group(*analyzeInput.GroupBy)
	}
	// time duration
	if len(analyzeInput.Duration) > 0 {
		t := analyzeInput.Duration[0]
		sql = sql.Where("created_at > ?", *t)
	}
	if len(analyzeInput.Duration) > 1 {
		t := analyzeInput.Duration[1]
		sql = sql.Where("created_at < ?", *t)
	}
	// limit
	if analyzeInput.Limit != nil {
		sql = sql.Limit(*analyzeInput.Limit)
	}
	// order by
	if analyzeInput.OrderBy != nil {
		sort := "asc"
		if analyzeInput.Sort != nil {
			sort = *analyzeInput.Sort
		}
		sql = sql.Order(fmt.Sprintf("%s %s", *analyzeInput.OrderBy, sort))
	}

	rows, err := sql.Rows()
	if err != nil {
		_ = rows.Close()
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeInternalError, err, "material_analysis")
	}

	_ = rows.Close()

	return nil, nil
}
