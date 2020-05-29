package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/logic"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"strings"
	"time"
)

func (r *queryResolver) AnalyzePoint(ctx context.Context, searchInput model.Search, limit int, offset int, pattern *string) (*model.PointResultsWrap, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("料号ID不能为空", "")
	}

	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("没有找到该尺寸所属的料号", "")
	}

	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
		return nil, NewGQLError("获取该料号的尺寸信息发生错误", err.Error())
	}

	pointSql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)
	if pattern != nil {
		pointSql = pointSql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *pattern))
	}

	var points []orm.Point
	if err := pointSql.Order("points.index ASC").Limit(limit).Offset(offset).Find(&points).Error; err != nil {
		return nil, NewGQLError("获取尺寸点位失败", err.Error())
	}

	var total int
	if err := pointSql.Count(&total).Error; err != nil {
		return nil, NewGQLError("统计检测点位总数时发生错误", err.Error())
	}

	end := searchInput.EndTime
	if end == nil {
		t := time.Now()
		end = &t
	}
	begin := searchInput.BeginTime
	if begin == nil {
		t := end.AddDate(0, -1, 0)
		begin = &t
	}

	var conds []string
	var vars []interface{}

	if searchInput.DeviceID != nil {
		conds = append(conds, "device_id = ?")
		vars = append(vars, *searchInput.DeviceID)
	}
	conds = append(conds, "created_at > ?")
	vars = append(vars, *begin)
	conds = append(conds, "created_at < ?")
	vars = append(vars, *end)

	if lineID, ok := searchInput.Extra["lineID"]; ok {
		conds = append(conds, "line_id = ?")
		vars = append(vars, lineID)
	}

	if mouldID, ok := searchInput.Extra["mouldID"]; ok {
		conds = append(conds, "mould_id = ?")
		vars = append(vars, mouldID)
	}

	if jigID, ok := searchInput.Extra["jigID"]; ok {
		conds = append(conds, "jig_id = ?")
		vars = append(vars, jigID)
	}

	if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
		conds = append(conds, "shift_number = ?")
		vars = append(vars, shiftNumber)
	}

	cond := strings.Join(conds, " AND ")
	var pointResults []*model.PointResult
	mpvs := make(map[int][]float64)
	sql := orm.DB.Model(&orm.PointValue{}).Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Select("point_values.v")

	for _, p := range points {
		var pointValues []float64
		rows, err := sql.Where("point_values.point_id = ?", p.ID).Rows()
		if err != nil {
			rows.Close()
			mpvs[p.ID] = pointValues
			continue
		}

		for rows.Next() {
			var v float64
			rows.Scan(&v)
			pointValues = append(pointValues, v)
		}
		rows.Close()
		mpvs[p.ID] = pointValues
	}

	for _, p := range points {
		data := mpvs[p.ID]
		s, cp, cpk, avg, ok, total, valueSet := logic.AnalyzePointValues(p, data)
		min, max, values, freqs, distribution := logic.Distribute(s, avg, valueSet)

		point := p
		pointResult := &model.PointResult{
			Total: &total,
			S:     &s,
			Ok:    &ok,
			Ng:    intP(total - ok),
			Cp:    &cp,
			Cpk:   &cpk,
			Avg:   &avg,
			Max:   &max,
			Min:   &min,
			Dataset: map[string]interface{}{
				"values":       values,
				"freqs":        freqs,
				"distribution": distribution,
			},
			Point: &model.Point{
				ID:         &point.ID,
				Name:       &point.Name,
				UpperLimit: &point.UpperLimit,
				Nominal:   &point.Nominal,
				LowerLimit: &point.LowerLimit,
			},
		}

		pointResults = append(pointResults, pointResult)
	}

	return &model.PointResultsWrap{
		PointResults: pointResults,
		Total:        total,
	}, nil
}

func (r *queryResolver) Sizes(ctx context.Context, page int, limit int, materialID int) (*model.SizeWrap, error) {
	var sizes []orm.Size
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Where("material_id = ?", materialID).Limit(limit).Offset(offset).Find(&sizes).Error; err != nil {
		return nil, NewGQLError("获取尺寸信息失败", err.Error())
	}
	var outs []*model.Size
	for _, v := range sizes {
		s := v
		outs = append(outs, &model.Size{
			ID:         &s.ID,
			Name:       &s.Name,
			MaterialID: &s.MaterialID,
		})
	}
	var count int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", materialID).Count(&count).Error; err != nil {
		return nil, NewGQLError("统计尺寸数量失败", err.Error())
	}

	return &model.SizeWrap{
		Total: &count,
		Sizes: outs,
	}, nil
}

func (r *queryResolver) TotalPointYield(ctx context.Context, searchInput model.Search, pattern *string) ([]*model.YieldWrap, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("料号ID不能为空", "")
	}

	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("没有找到该尺寸所属的料号", "")
	}

	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
		return nil, NewGQLError("获取该料号的尺寸信息发生错误", err.Error())
	}

	pointSql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)
	if pattern != nil {
		pointSql = pointSql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *pattern))
	}

	var points []orm.Point
	if err := pointSql.Order("points.index ASC").Find(&points).Error; err != nil {
		return nil, NewGQLError("获取尺寸点位失败", err.Error())
	}

	end := searchInput.EndTime
	if end == nil {
		t := time.Now()
		end = &t
	}
	begin := searchInput.BeginTime
	if begin == nil {
		t := end.AddDate(0, -1, 0)
		begin = &t
	}

	var conds []string
	var vars []interface{}

	if searchInput.DeviceID != nil {
		conds = append(conds, "device_id = ?")
		vars = append(vars, *searchInput.DeviceID)
	}
	conds = append(conds, "created_at > ?")
	vars = append(vars, *begin)
	conds = append(conds, "created_at < ?")
	vars = append(vars, *end)

	if lineID, ok := searchInput.Extra["lineID"]; ok {
		conds = append(conds, "line_id = ?")
		vars = append(vars, lineID)
	}

	if mouldID, ok := searchInput.Extra["mouldID"]; ok {
		conds = append(conds, "mould_id = ?")
		vars = append(vars, mouldID)
	}

	if jigID, ok := searchInput.Extra["jigID"]; ok {
		conds = append(conds, "jig_id = ?")
		vars = append(vars, jigID)
	}

	if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
		conds = append(conds, "shift_number = ?")
		vars = append(vars, shiftNumber)
	}

	cond := strings.Join(conds, " AND ")
	mpvs := make(map[int][]float64)
	sql := orm.DB.Model(&orm.PointValue{}).Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Select("point_values.v")

	for _, p := range points {
		var pointValues []float64
		rows, err := sql.Where("point_values.point_id = ?", p.ID).Rows()
		if err != nil {
			rows.Close()
			mpvs[p.ID] = pointValues
			continue
		}

		for rows.Next() {
			var v float64
			rows.Scan(&v)
			pointValues = append(pointValues, v)
		}
		rows.Close()
		mpvs[p.ID] = pointValues
	}

	var out []*model.YieldWrap
	for _, p := range points {
		data := mpvs[p.ID]
		total := len(data)
		ok := 0
		for _, v := range data {
			if p.NotValid(v) {
				continue
			}

			if v >= p.LowerLimit && v <= p.UpperLimit {
				ok++
			}
		}
		o := &model.YieldWrap{
			Name:  p.Name,
			Value: float64(ok) / float64(total),
		}
		out = append(out, o)
	}

	return out, nil
}
