package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/jinzhu/copier"
	"strconv"
	"strings"
	"time"
)

func Point(ctx context.Context, id int) (*model.Point, error) {
	var point orm.Point
	if err := point.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "point")
	}

	var out model.Point
	if err := copier.Copy(&out, &point); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "point")
	}

	return &out, nil
}

// 尺寸不良率排行
func SizeUnYieldTop(ctx context.Context, groupInput model.GraphInput, versionID *int) (*model.EchartsResult, error) {
	var material orm.Material
	if err := material.Get(uint(groupInput.TargetID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}
	var version orm.MaterialVersion
	if versionID != nil {
		if err := orm.Model(&orm.MaterialVersion{}).Where("id = ?", *versionID).Find(&version).Error; err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
		}
	} else {
		v, err := material.GetCurrentVersion()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeActiveVersionNotFound, err, "material_version")
		}

		version = *v
	}

	query := orm.DB.Model(&orm.Product{}).Where("products.material_id = ?", groupInput.TargetID)
	query = query.Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("import_records.blocked = ? AND products.material_version_id = ?", false, version.ID)

	// filters
	if groupInput.Filters != nil {
		for k, v := range groupInput.Filters {
			switch k {
			case "device_id":
				query = query.Where("products.device_id = ?", v)
			case "shift":
				switch fmt.Sprint(v) {
				case "A":
					query = query.Where("TIME(products.created_at) >= '08:00:00' AND TIME(products.created_at) <= '17:30:00'")
				case "B":
					query = query.Where("TIME(products.created_at) < '08:00:00' OR TIME(products.created_at) > '17:30:00'")
				}
			default:
				query = query.Where(fmt.Sprintf("JSON_EXTRACT(products.attribute, '$.\"%s\"') = ?", k), v)
			}
		}
	}

	// duration
	if len(groupInput.Duration) > 0 {
		query = query.Where("products.created_at > ?", groupInput.Duration[0])
	}
	if len(groupInput.Duration) > 1 {
		query = query.Where("products.created_at < ?", groupInput.Duration[1])
	}

	var products []orm.Product
	if err := query.Find(&products).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "size")
	}

	var points []orm.Point
	if err := orm.Model(&orm.Point{}).Where("material_id = ?", groupInput.TargetID).Find(&points).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "points")
	}

	var xAxisData []string
	var data []float64
	var amount []int
	var total = len(products)

	if total == 0 {
		return &model.EchartsResult{
			XAxisData:        []string{},
			SeriesData:       types.Map{"data": []float64{}},
			SeriesAmountData: types.Map{"data": []int{}},
		}, nil
	}

	for _, point := range points {
		var ok int
		for _, product := range products {
			v := product.PointValues[point.Name]
			value, err := strconv.ParseFloat(fmt.Sprint(v), 64)
			if err != nil {
				continue
			}

			if value <= point.UpperLimit && value >= point.LowerLimit {
				ok++
			}
		}
		rate := float64(total-ok) / float64(total)
		xAxisData = append(xAxisData, point.Name)
		data = append(data, rate)
		amount = append(amount, total-ok)
	}

	// 排序
	sort := model.SortDesc
	if groupInput.Sort != nil {
		sort = *groupInput.Sort
	}

	length := len(xAxisData)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1-i; j++ {
			if (data[j] <= data[j+1] && sort == model.SortAsc) || (data[j] >= data[j+1] && sort == model.SortDesc) {
				continue
			}
			data[j], data[j+1] = data[j+1], data[j]
			xAxisData[j], xAxisData[j+1] = xAxisData[j+1], xAxisData[j]
			amount[j], amount[j+1] = amount[j+1], amount[j]
		}
	}

	// limit
	if groupInput.Limit != nil && *groupInput.Limit < len(points) {
		limit := *groupInput.Limit
		data = data[:limit]
		xAxisData = xAxisData[:limit]
		amount = amount[:limit]
	}

	return &model.EchartsResult{
		XAxisData:        xAxisData,
		SeriesData:       types.Map{"data": data},
		SeriesAmountData: types.Map{"data": amount},
	}, nil
}

func PointList(ctx context.Context, materialID int, versionID *int, search *string, limit int, page int) (*model.PointListWithYieldResponse, error) {
	var version orm.MaterialVersion
	if versionID != nil {
		if err := version.Get(uint(*versionID)); err != nil {
			return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material_version")
		}
	} else {
		err := orm.Model(&orm.MaterialVersion{}).Where("material_id = ? AND active = ?", materialID, true).Find(&version).Error
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
		}
	}

	sql := orm.Model(&orm.Point{}).Where("material_id = ? AND material_version_id = ?", materialID, version.ID)
	if search != nil {
		sql = sql.Where("name like ?", fmt.Sprintf("%%%s%%", *search))
	}

	var total int
	if err := sql.Count(&total).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "point")
	}

	var points []orm.Point
	if err := sql.Limit(limit).Offset((page - 1) * limit).Find(&points).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "point")
	}

	//var products []orm.Product
	//var now = time.Now()
	//now = now.AddDate(0, -1, 0)
	//query := orm.Model(&orm.Product{}).Where("products.material_id = ? AND products.created_at > ?", materialID, now)
	//query = query.Joins("JOIN import_records ON import_records.id = products.import_record_id")
	//query = query.Where("import_records.blocked = ? AND products.material_version_id = ?", false, version.ID)
	//t := time.Now()
	//t.AddDate(0, 0, -7)
	//query = query.Where("products.created_at > ?", t)
	//
	//if err := query.Find(&products).Error; err != nil {
	//	return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
	//}

	//var sum = len(products)
	var outs []*model.Point
	for _, p := range points {
		//var ok int
		//var total = sum
		//for _, product := range products {
		//	v, exist := product.PointValues[p.Name]
		//	if !exist {
		//		total--
		//		continue
		//	}
		//
		//	pointValue, err := strconv.ParseFloat(fmt.Sprint(v), 64)
		//	if err != nil {
		//		log.Errorln(err)
		//		total--
		//		continue
		//	}
		//
		//	if pointValue > p.LowerLimit && pointValue < p.UpperLimit {
		//		ok++
		//	}
		//}

		var point model.Point
		copier.Copy(&point, &p)
		outs = append(outs, &point)
	}

	return &model.PointListWithYieldResponse{
		Total: total,
		List:  outs,
	}, nil
}

func SizeNormalDistribution(ctx context.Context, id int, duration []*time.Time, filters map[string]interface{}) (*model.PointResult, error) {
	var point orm.Point
	if err := point.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "point")
	}

	query := orm.Model(&orm.Product{}).Select(fmt.Sprintf("JSON_EXTRACT(products.`point_values`, '$.\"%s\"') AS point_value", point.Name))
	query = query.Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("import_records.blocked = ?", false)
	query = query.Where("products.material_id = ?", point.MaterialID)
	if len(duration) > 0 {
		query = query.Where("products.created_at > ?", duration[0])
	}

	if len(duration) > 1 {
		query = query.Where("products.created_at < ?", duration[1])
	}

	for key, value := range filters {
		switch key {
		case "device_id":
			query = query.Where("products.device_id = ?", fmt.Sprint(value))
		case "shift":
			switch fmt.Sprint(value) {
			case "A":
				query = query.Where("TIME(products.created_at) >= '08:00:00' AND TIME(products.created_at) <= '17:30:00'")
			case "B":
				query = query.Where("TIME(products.created_at) < '08:00:00' OR TIME(products.created_at) > '17:30:00'")
			}
		default:
			query = query.Where(fmt.Sprintf("JSON_EXTRACT(products.attribute, '$.\"%s\"') = ?", key), value)
		}
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "point_values")
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

	s, cp, cpk, avg, ok, total, valueSet := AnalyzePointValues(point, data)
	fmt.Println("----------------------- ok", ok)
	min, max, values, freqs, distribution := Distribute(s, avg, valueSet)

	var outPoint model.Point
	if err := copier.Copy(&outPoint, &point); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "point")
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

func GroupAnalyzePoint(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	var point orm.Point
	if err := point.Get(uint(analyzeInput.TargetID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "point")
	}

	query := orm.DB.Model(orm.Product{}).Where("products.material_id = ?", point.MaterialID)
	query = query.Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("import_records.blocked = ?", false)

	var selectQueries, groupColumns []string
	var selectVariables []interface{}
	var joinDevice = false

	// amount
	selectQueries = append(selectQueries, "COUNT(products.id) AS amount")

	// axis
	selectQueries = append(selectQueries, "%v AS axis")
	groupColumns = append(groupColumns, "axis")
	switch analyzeInput.XAxis {
	case model.CategoryDate:
		selectVariables = append(selectVariables, "DATE(products.created_at)")
	case model.CategoryDevice:
		selectVariables = append(selectVariables, "devices.name")
		joinDevice = true
	case model.CategoryShift:
		// TODO: 解决UTC时区问题
		// 暂时采用UTC 00:00:00 - 09:30:00
		selectVariables = append(selectVariables, "TIME(products.created_at) >= '00:00:00' && TIME(products.created_at) <= '09:30:00'")
	default:
		if analyzeInput.AttributeXAxis == nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("need AttributeXAxis when xAxis type is attribute"))
		}
		selectVariables = append(selectVariables, fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(products.`attribute`, '$.\"%v\"'))", *analyzeInput.AttributeXAxis))
	}

	// group by
	if analyzeInput.GroupBy != nil {
		selectQueries = append(selectQueries, "%v as group_by")
		groupColumns = append(groupColumns, "group_by")
		switch *analyzeInput.GroupBy {
		case model.CategoryDate:
			selectVariables = append(selectVariables, "DATE(products.created_at)")
		case model.CategoryDevice:
			selectVariables = append(selectVariables, "devices.name")
			joinDevice = true
		case model.CategoryShift:
			selectVariables = append(selectVariables, "TIME(products.created_at) >= '00:00:00' && TIME(products.created_at) <= '09:30:00'")
		default:
			if analyzeInput.AttributeGroup == nil {
				return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("need AttributeGroup when groupBy type is attribute"))
			}
			selectVariables = append(selectVariables, fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(products.`attribute`, '$.\"%v\"'))", *analyzeInput.AttributeGroup))
		}
	}

	// join device
	if joinDevice {
		query = query.Joins("JOIN devices ON products.device_id = devices.id")
	}

	// assemble selects
	query = query.Select(fmt.Sprintf(strings.Join(selectQueries, ", "), selectVariables...))

	// assemble groups
	query = query.Group(strings.Join(groupColumns, ", "))

	// time duration
	if len(analyzeInput.Duration) > 0 {
		t := analyzeInput.Duration[0]
		query = query.Where("products.created_at > ?", *t)
	}
	if len(analyzeInput.Duration) > 1 {
		t := analyzeInput.Duration[1]
		query = query.Where("products.created_at < ?", *t)
	}

	sort := model.SortAsc
	if analyzeInput.Sort != nil {
		sort = *analyzeInput.Sort
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "point_values")
	}

	results := scanRows(rows, analyzeInput.GroupBy)

	if analyzeInput.YAxis == "UnYield" {
		query = query.Where(
			fmt.Sprintf("JSON_EXTRACT(products.`point_values`, '$.\"%s\"') < ? OR JSON_EXTRACT(products.`point_values`, '$.\"%s\"') > ?", point.Name, point.Name),
			point.LowerLimit, point.UpperLimit,
		)
	} else {
		query = query.Where(
			fmt.Sprintf("JSON_EXTRACT(products.`point_values`, '$.\"%s\"') >= ? AND JSON_EXTRACT(products.`point_values`, '$.\"%s\"') <= ?", point.Name, point.Name),
			point.LowerLimit, point.UpperLimit,
		)
	}

	rows, err = query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "point_values")
	}
	qualifiedResults := scanRows(rows, analyzeInput.GroupBy)

	return calYieldAnalysisResult(results, qualifiedResults, analyzeInput.Limit, sort.String())
}
