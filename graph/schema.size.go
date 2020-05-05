package graph

import (
	"context"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"strings"
	"time"
)

func (r *queryResolver) AnalyzePoint(ctx context.Context, searchInput model.Search, limit int, offset int) (*model.PointResultsWrap, error) {
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

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index ASC").Limit(limit).Offset(offset).Find(&points).Error; err != nil {
		return nil, NewGQLError("获取尺寸点位失败", err.Error())
	}

	var total int
	if err := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Count(&total).Error; err != nil {
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
	cond := strings.Join(conds, " AND ")

	var pointResults []*model.PointResult

	for _, p := range points {
		var pointValues []orm.PointValue

		err := orm.DB.Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Where("point_values.point_id = ?", p.ID).Find(&pointValues).Error
		if err != nil {
			return nil, NewGQLError("获取数据失败", err.Error())
		}

		total := len(pointValues)
		ok := 0
		valueSet := make([]float64, 0)
		for _, v := range pointValues {
			if p.NotValid(v.V) {
				continue
			}

			valueSet = append(valueSet, v.V)
			if v.V >= p.LowerLimit && v.V <= p.UpperLimit {
				ok++
			}
		}

		s := logic.RMSError(valueSet)
		cp := logic.Cp(p.UpperLimit, p.LowerLimit, s)
		avg := logic.Average(valueSet)
		cpk := logic.Cpk(p.UpperLimit, p.LowerLimit, avg, s)
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
				Norminal:   &point.Norminal,
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
