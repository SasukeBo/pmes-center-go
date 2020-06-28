package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
	"strings"
	"time"
)

func (r *queryResolver) SizeNormalDistribution(ctx context.Context, id int, duration []*time.Time, filters map[string]interface{}) (*model.PointResult, error) {
	var point orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("id = ?", id).Find(&point).Error; err != nil {
		return nil, NewGQLError("获取尺寸信息失败。", err.Error())
	}
	sql := orm.DB.Model(&orm.PointValue{}).Select("point_values.v").Joins(
		"LEFT JOIN products ON point_values.product_uuid = products.uuid",
	).Where("point_values.point_id = ?", id)
	rows, err := sql.Rows()
	if err != nil {
		return nil, NewGQLError("获取数据失败，发生错误", err.Error())
	}
	var data []float64
	for rows.Next() {
		for rows.Next() {
			var v float64
			rows.Scan(&v)
			data = append(data, v)
		}
		rows.Close()
	}

	s, cp, cpk, avg, ok, total, valueSet := logic.AnalyzePointValues(point, data)
	fmt.Println("----------------------- ok", ok)
	min, max, values, freqs, distribution := logic.Distribute(s, avg, valueSet)

	var outPoint model.Point
	if err := copier.Copy(&outPoint, &point); err != nil {
		return nil, NewGQLError("转换数据时发生错误", err.Error())
	}

	return &model.PointResult{
		Total: total,
		S:     s,
		Ok:    ok,
		Ng:    total - ok,
		Cp:    cp,
		Cpk:   cpk,
		Avg:   avg,
		Max:   max,
		Min:   min,
		Point: &outPoint,
		Dataset: map[string]interface{}{
			"values":       values,
			"freqs":        freqs,
			"distribution": distribution,
		},
	}, nil
}

//func (r *queryResolver) AnalyzePoint(ctx context.Context, searchInput model.Search, limit int, offset int, pattern *string) (*model.PointResultsWrap, error) {
//	if searchInput.MaterialID == nil {
//		return nil, NewGQLError("料号ID不能为空", "")
//	}
//
//	material := orm.GetMaterialWithID(*searchInput.MaterialID)
//	if material == nil {
//		return nil, NewGQLError("没有找到该尺寸所属的料号", "")
//	}
//
//	var sizeIDs []int
//	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
//		return nil, NewGQLError("获取该料号的尺寸信息发生错误", err.Error())
//	}
//
//	pointSql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)
//	if pattern != nil {
//		pointSql = pointSql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *pattern))
//	}
//
//	var points []orm.Point
//	if err := pointSql.Order("points.index ASC").Limit(limit).Offset(offset).Find(&points).Error; err != nil {
//		return nil, NewGQLError("获取尺寸点位失败", err.Error())
//	}
//
//	var total int
//	if err := pointSql.Count(&total).Error; err != nil {
//		return nil, NewGQLError("统计检测点位总数时发生错误", err.Error())
//	}
//
//	end := searchInput.EndTime
//	if end == nil {
//		t := time.Now()
//		end = &t
//	}
//	begin := searchInput.BeginTime
//	if begin == nil {
//		t := end.AddDate(0, -1, 0)
//		begin = &t
//	}
//
//	var conds []string
//	var vars []interface{}
//
//	if searchInput.DeviceID != nil {
//		conds = append(conds, "device_id = ?")
//		vars = append(vars, *searchInput.DeviceID)
//	}
//	conds = append(conds, "created_at > ?")
//	vars = append(vars, *begin)
//	conds = append(conds, "created_at < ?")
//	vars = append(vars, *end)
//
//	if lineID, ok := searchInput.Extra["lineID"]; ok {
//		conds = append(conds, "line_id = ?")
//		vars = append(vars, lineID)
//	}
//
//	if mouldID, ok := searchInput.Extra["mouldID"]; ok {
//		conds = append(conds, "mould_id = ?")
//		vars = append(vars, mouldID)
//	}
//
//	if jigID, ok := searchInput.Extra["jigID"]; ok {
//		conds = append(conds, "jig_id = ?")
//		vars = append(vars, jigID)
//	}
//
//	if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
//		conds = append(conds, "shift_number = ?")
//		vars = append(vars, shiftNumber)
//	}
//
//	cond := strings.Join(conds, " AND ")
//	var pointResults []*model.PointResult
//	mpvs := make(map[int][]float64)
//	sql := orm.DB.Model(&orm.PointValue{}).Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Select("point_values.v")
//
//	for _, p := range points {
//		var pointValues []float64
//		rows, err := sql.Where("point_values.point_id = ?", p.ID).Rows()
//		if err != nil {
//			rows.Close()
//			mpvs[p.ID] = pointValues
//			continue
//		}
//
//		for rows.Next() {
//			var v float64
//			rows.Scan(&v)
//			pointValues = append(pointValues, v)
//		}
//		rows.Close()
//		mpvs[p.ID] = pointValues
//	}
//
//	for _, p := range points {
//		data := mpvs[p.ID]
//		s, cp, cpk, avg, ok, total, valueSet := logic.AnalyzePointValues(p, data)
//		min, max, values, freqs, distribution := logic.Distribute(s, avg, valueSet)
//
//		point := p
//		pointResult := &model.PointResult{
//			Total: &total,
//			S:     &s,
//			Ok:    &ok,
//			Ng:    intP(total - ok),
//			Cp:    &cp,
//			Cpk:   &cpk,
//			Avg:   &avg,
//			Max:   &max,
//			Min:   &min,
//			Dataset: map[string]interface{}{
//				"values":       values,
//				"freqs":        freqs,
//				"distribution": distribution,
//			},
//			Point: &model.Point{
//				ID:         &point.ID,
//				Name:       &point.Name,
//				UpperLimit: &point.UpperLimit,
//				Nominal:    &point.Nominal,
//				LowerLimit: &point.LowerLimit,
//			},
//		}
//
//		pointResults = append(pointResults, pointResult)
//	}
//
//	return &model.PointResultsWrap{
//		PointResults: pointResults,
//		Total:        total,
//	}, nil
//}

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
		ng := 0
		for _, v := range data {
			if p.NotValid(v) {
				continue
			}

			if v < p.LowerLimit || v > p.UpperLimit {
				ng++
			}
		}
		o := &model.YieldWrap{
			Name:  p.Name,
			Ng:    ng,
			Value: float64(ng) / float64(total),
		}
		out = append(out, o)
	}

	length := len(out)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1-i; j++ {
			if out[j].Value < out[j+1].Value {
				out[j], out[j+1] = out[j+1], out[j]
			}
		}
	}

	if length > 20 {
		return out[:20], nil
	}

	return out, nil
}

func (r *queryResolver) PointListWithYield(ctx context.Context, materialID int, limit int, page int) (*model.PointListWithYieldResponse, error) {
	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", materialID).Pluck("id", &sizeIDs).Error; err != nil {
		return nil, NewGQLError("获取尺寸失败", err.Error())
	}

	sql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, NewGQLError("统计尺寸数量发生错误", err.Error())
	}

	var points []orm.Point
	if err := sql.Limit(limit).Offset((page - 1) * limit).Find(&points).Error; err != nil {
		return nil, NewGQLError("获取尺寸失败", err.Error())
	}

	var list []*model.PointYield
	for _, p := range points {
		var total, ok int
		query := orm.DB.Model(&orm.PointValue{}).Where("point_id = ?", p.ID)
		query.Count(&total)
		query.Where("v > ? AND v < ?", p.LowerLimit, p.UpperLimit).Count(&ok)
		var point model.NewPoint
		copier.Copy(&point, &p)

		list = append(list, &model.PointYield{
			Point: &point,
			Ok:    ok,
			Total: total,
		})
	}

	return &model.PointListWithYieldResponse{
		Total: total,
		List:  list,
	}, nil
}

// 尺寸不良率排行
func (r *queryResolver) SizeUnYieldTop(ctx context.Context, groupInput model.GroupAnalyzeInput) (*model.EchartsResult, error) {
	query := orm.DB.Model(&orm.Product{}).Where("products.material_id = ?", groupInput.TargetID)

	// 连接 point values
	query = query.Joins("JOIN point_values AS pv ON pv.product_uuid = products.uuid")

	// 连接 point
	query = query.Joins("JOIN points AS p ON pv.point_id = p.id")

	// select fields
	query = query.Select("p.name AS point_name, (pv.v >= p.lower_limit && pv.v <= p.upper_limit) AS qualified")

	// duration
	if len(groupInput.Duration) > 0 {
		query = query.Where("products.created_at > ?", groupInput.Duration[0])
	}
	if len(groupInput.Duration) > 1 {
		query = query.Where("products.created_at < ?", groupInput.Duration[1])
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, NewGQLError("查询数据发生错误", err.Error())
	}

	type result struct {
		Name      string
		Qualified bool
	}
	var rs []result
	for rows.Next() {
		var r result
		if err := rows.Scan(&r.Name, &r.Qualified); err != nil {
			continue
		}
		rs = append(rs, r)
	}

	type summary struct {
		OK int
		NG int
	}
	summaryMap := make(map[string]summary)
	for _, o := range rs {
		s, ok := summaryMap[o.Name]
		if !ok {
			s = summary{}
		}
		if o.Qualified {
			s.OK++
		} else {
			s.NG++
		}
		summaryMap[o.Name] = s
	}

	var points []string
	var unYields []float64
	for k, s := range summaryMap {
		points = append(points, k)
		var yield float64
		if s.NG+s.OK == 0 {
			yield = 0
		} else {
			yield = float64(s.NG) / float64(s.OK+s.NG)
		}
		unYields = append(unYields, yield)
	}

	sort := model.SortDesc
	if groupInput.Sort != nil {
		sort = *groupInput.Sort
	}

	length := len(points)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1-i; j++ {
			if (unYields[j] <= unYields[j+1] && sort == model.SortAsc) || (unYields[j] >= unYields[j+1] && sort == model.SortDesc) {
				continue
			}
			unYields[j], unYields[j+1] = unYields[j+1], unYields[j]
			points[j], points[j+1] = points[j+1], points[j]
		}
	}

	if groupInput.Limit != nil && *groupInput.Limit < len(points) {
		points = points[:*groupInput.Limit]
		unYields = unYields[:*groupInput.Limit]
	}

	return &model.EchartsResult{
		XAxisData: points,
		SeriesData: map[string]interface{}{
			"data": unYields,
		},
	}, nil
}

func (r *queryResolver) GroupAnalyzeSize(ctx context.Context, analyzeInput model.GroupAnalyzeInput) (*model.EchartsResult, error) {
	query := orm.DB.Model(orm.PointValue{}).Where("point_values.point_id = ?", analyzeInput.TargetID)

	var selectQueries, groupColumns []string
	var selectVariables []interface{}
	var joinDevice = false
	var joins = []string{
		"JOIN points AS p ON point_values.point_id = p.id",
		"JOIN products AS pd ON point_values.product_uuid = pd.uuid",
	}

	// amount
	selectQueries = append(selectQueries, "COUNT(point_values.product_uuid) AS amount")

	// axis
	selectQueries = append(selectQueries, "%v AS axis")
	groupColumns = append(groupColumns, "axis")
	switch analyzeInput.XAxis {
	case model.CategoryDate:
		selectVariables = append(selectVariables, "DATE(pd.created_at)")
	case model.CategoryDevice:
		selectVariables = append(selectVariables, "devices.name")
		joinDevice = true
	default:
		selectVariables = append(selectVariables, fmt.Sprintf("pd.%v", analyzeInput.XAxis))
	}

	// group by
	if analyzeInput.GroupBy != nil {
		selectQueries = append(selectQueries, "%v as group_by")
		groupColumns = append(groupColumns, "group_by")
		switch *analyzeInput.GroupBy {
		case model.CategoryDate:
			selectVariables = append(selectVariables, "DATE(created_at)")
		case model.CategoryDevice:
			selectVariables = append(selectVariables, "devices.name")
			joinDevice = true
		default:
			selectVariables = append(selectVariables, fmt.Sprintf("pd.%v", *analyzeInput.GroupBy))
		}
	}

	// join device
	if joinDevice {
		joins = append(joins, "JOIN devices ON pd.device_id = devices.id")
	}
	query = query.Joins(strings.Join(joins, " "))

	// assemble selects
	query = query.Select(fmt.Sprintf(strings.Join(selectQueries, ", "), selectVariables...))

	// assemble groups
	query = query.Group(strings.Join(groupColumns, ", "))

	// time duration
	if len(analyzeInput.Duration) > 0 {
		t := analyzeInput.Duration[0]
		query = query.Where("created_at > ?", *t)
	}
	if len(analyzeInput.Duration) > 1 {
		t := analyzeInput.Duration[1]
		query = query.Where("created_at < ?", *t)
	}

	sort := model.SortAsc
	if analyzeInput.Sort != nil {
		sort = *analyzeInput.Sort
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, NewGQLError("分析数据时发生错误", err.Error())
	}

	results := scanRows(rows, analyzeInput.GroupBy)

	if analyzeInput.YAxis == "UnYield" {
		query = query.Where("point_values.v < p.lower_limit OR point_values.v > p.upper_limit")
	} else {
		query = query.Where("point_values.v >= p.lower_limit AND point_values.v <= p.upper_limit")
	}

	rows, err = query.Rows()
	if err != nil {
		return nil, NewGQLError("分析数据发生错误", err.Error())
	}
	qualifiedResults := scanRows(rows, analyzeInput.GroupBy)

	return calYieldAnalysisResult(results, qualifiedResults, analyzeInput.Limit, sort.String())
}
