package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/copier"
	"strings"
	"time"
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

type analysis struct {
	Amount  int64
	Axis    string
	GroupBy string
}

func AnalyzeMaterial(ctx context.Context, analyzeInput model.AnalyzeMaterialInput) (*model.EchartsResult, error) {
	query := orm.Model(&orm.Product{}).Where("products.material_id = ?", analyzeInput.MaterialID)
	var selectQueries, groupColumns []string
	var selectVariables []interface{}
	var joinDevice = false

	// amount
	selectQueries = append(selectQueries, "COUNT(products.id) as amount")

	// axis
	selectQueries = append(selectQueries, "%v as axis")
	groupColumns = append(groupColumns, "axis")
	switch analyzeInput.XAxis {
	case model.CategoryDate:
		selectVariables = append(selectVariables, "DATE(created_at)")
	case model.CategoryDevice:
		selectVariables = append(selectVariables, "devices.name")
		joinDevice = true
	case model.CategoryAttribute:
		if analyzeInput.AttributeXAxis == nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAnalyzeMissingAttributeXAxis, nil)
		}
		selectVariables = append(selectVariables, fmt.Sprintf("JSON_EXTRACT(`attribute`, '$.\"%s\"')", *analyzeInput.AttributeXAxis))
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
		case model.CategoryAttribute:
			if analyzeInput.AttributeGroup == nil {
				return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAnalyzeMissingAttributeGroup, nil)
			}
			selectVariables = append(selectVariables, fmt.Sprintf("JSON_EXTRACT(`attribute`, '$.\"%s\"')", *analyzeInput.AttributeGroup))
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
		query = query.Where("created_at > ?", *t)
	}
	if len(analyzeInput.Duration) > 1 {
		t := analyzeInput.Duration[1]
		query = query.Where("created_at < ?", *t)
	}
	// order by xAxis
	sort := "asc"
	if analyzeInput.Sort != nil {
		v := *analyzeInput.Sort
		sort = string(v)
	}
	query = query.Order(fmt.Sprintf("axis %s", sort))

	rows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAnalyzeError, err)
	}

	results := scanRows(rows, analyzeInput.GroupBy != nil)

	if analyzeInput.Limit != nil && *analyzeInput.Limit < 0 {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAnalyzeIllegalInput, errors.New("limit cannot below 0"))
	}

	if analyzeInput.YAxis == "Yield" {
		rows, err := query.Where("qualified = ?", true).Rows()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeMaterialAnalyzeError, err)
		}
		qualifiedResults := scanRows(rows, analyzeInput.GroupBy != nil)

		return calYieldAnalysisResult(results, qualifiedResults, analyzeInput.Limit)
	}

	return calAmountAnalysisResult(results, analyzeInput.Limit)
}

func calYieldAnalysisResult(results, qualifiedResults []analysis, limit *int) (*model.EchartsResult, error) {
	totalAmount, _ := calAmountAnalysisResult(results, limit)
	yieldAmount, _ := calAmountAnalysisResult(qualifiedResults, limit)

	for i, item := range totalAmount.XAxisData {
		var index = findIndex(yieldAmount.XAxisData, item)
		for k, v := range totalAmount.SeriesData {
			data := v.([]interface{})
			if index < 0 {
				data[i] = 0
			} else {
				total := data[i].(int)
				yieldData := yieldAmount.SeriesData[k].([]interface{})
				yield := yieldData[index].(int)
				value := float64(yield) / float64(total)
				data[i] = value
			}
			totalAmount.SeriesData[k] = data
		}
	}

	return totalAmount, nil
}

func findIndex(list []string, find string) int {
	for i, v := range list {
		if v == find {
			return i
		}
	}

	return -1
}

func calAmountAnalysisResult(scanResults []analysis, limit *int) (*model.EchartsResult, error) {
	var xAxisMapData = make(map[string]int)
	var xAxisData []string
	var seriesMapData = make(map[string]interface{})
	for i, result := range scanResults {
		xAxis := fmt.Sprint(result.Axis)
		if _, ok := xAxisMapData[xAxis]; !ok {
			xAxisMapData[xAxis] = i
			xAxisData = append(xAxisData, xAxis)
		}
		if data, ok := seriesMapData[fmt.Sprint(result.GroupBy)]; ok {
			seriesMap := data.(map[string]interface{})
			seriesMap[fmt.Sprint(result.Axis)] = int(result.Amount)
			seriesMapData[fmt.Sprint(result.GroupBy)] = seriesMap
		} else {
			seriesMap := map[string]interface{}{fmt.Sprint(result.Axis): int(result.Amount)}
			seriesMapData[fmt.Sprint(result.GroupBy)] = seriesMap
		}
	}

	if limit != nil {
		xAxisData = xAxisData[:*limit]
	}

	var seriesData = make(map[string]interface{})
	for _, item := range xAxisData {
		for k, v := range seriesMapData {
			sdv, ok := seriesData[k]
			var dataSet []interface{}
			if ok {
				dataSet = sdv.([]interface{})
			} else {
				dataSet = make([]interface{}, 0)
			}

			seriesMap := v.(map[string]interface{})
			var value interface{}
			if v, ok := seriesMap[item]; ok {
				value = v
			} else {
				value = 0
			}

			dataSet = append(dataSet, value)
			seriesData[k] = dataSet
		}
	}

	return &model.EchartsResult{
		XAxisData:  xAxisData,
		SeriesData: seriesData,
	}, nil
}

func scanRows(rows *sql.Rows, needGroup bool) []analysis {
	var results []analysis
	for rows.Next() {
		var result = analysis{GroupBy: "data"}
		var err error
		if needGroup {
			err = rows.Scan(&result.Amount, &result.Axis, &result.GroupBy)
		} else {
			err = rows.Scan(&result.Amount, &result.Axis)
		}
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	_ = rows.Close()

	return results
}

func MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	query := orm.Model(&orm.Product{}).Select(
		"materials.name AS name, COUNT(products.id) AS amount",
	).Joins(
		"JOIN materials ON products.material_id = materials.id",
	).Group("products.material_id")

	if len(duration) > 0 {
		t := duration[0]
		query = query.Where("products.created_at > ?", *t)
	}

	if len(duration) > 1 {
		t := duration[1]
		query = query.Where("products.created_at < ?", *t)
	}

	var totalResult = make(map[string]int)
	totalRows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "products")
	}

	for totalRows.Next() {
		var name string
		var amount int64
		err := totalRows.Scan(&name, &amount)
		if err != nil {
			continue
		}

		totalResult[name] = int(amount)
	}
	fmt.Println(totalResult)

	var ngResult = make(map[string]int)
	ngRows, err := query.Where("qualified = ?", false).Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeCountObjectFailed, err, "products")
	}

	for ngRows.Next() {
		var name string
		var amount int64
		err := ngRows.Scan(&name, &amount)
		if err != nil {
			continue
		}

		ngResult[name] = int(amount)
	}
	fmt.Println(ngResult)

	var seriesData []float64
	var xAxisData []string

	for k, total := range totalResult {
		xAxisData = append(xAxisData, k)
		var rate float64 = 0
		if ng, ok := ngResult[k]; ok {
			rate = float64(ng) / float64(total)
		}
		seriesData = append(seriesData, rate)
	}

	var length = len(seriesData)
	for i := 0; i < length-1; i++ {
		for j := 0; j < length-1-i; j++ {
			if seriesData[j] < seriesData[j+1] {
				s := seriesData[j]
				x := xAxisData[j]
				seriesData[j] = seriesData[j+1]
				xAxisData[j] = xAxisData[j+1]
				seriesData[j+1] = s
				xAxisData[j+1] = x
			}
		}
	}

	if limit > len(xAxisData) {
		limit = len(xAxisData)
	}

	return &model.EchartsResult{
		XAxisData: xAxisData[:limit],
		SeriesData: map[string]interface{}{
			"data": seriesData[:limit],
		},
	}, nil
}
