package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"strings"
	"time"
)

type analysis struct {
	Amount  int64
	Axis    string
	GroupBy string
}

type echartsResult struct {
	XAxisData        []string
	SeriesData       map[string][]float64
	SeriesAmountData map[string][]float64
}

func groupAnalyze(ctx context.Context, params model.GraphInput, target string, versionID *int) (*model.EchartsResult, error) {
	t1 := time.Now()
	query := orm.DB.Model(&orm.Product{})
	switch target {
	case "material":
		query = query.Where("products.material_id = ?", params.TargetID)
	case "device":
		query = query.Where("products.device_id = ?", params.TargetID)
	}
	var selectQueries, groupColumns, joins []string
	var selectVariables []interface{}
	var joinDevice = false

	// 连接 import_records
	joins = append(joins, "JOIN import_records ON products.import_record_id = import_records.id")
	query = query.Where("import_records.blocked = ?", false)

	// 连接 material_version
	if versionID != nil {
		query = query.Where("products.material_version_id  = ?", *versionID)
	} else {
		joins = append(joins, "JOIN material_versions ON products.material_version_id = material_versions.id")
		query = query.Where("material_versions.active  = ?", true)
	}

	// amount
	selectQueries = append(selectQueries, "COUNT(products.id) as amount")

	// axis
	selectQueries = append(selectQueries, "%v as axis")
	groupColumns = append(groupColumns, "axis")
	switch params.XAxis {
	case model.CategoryDate:
		selectVariables = append(selectVariables, "DATE(products.created_at)")
	case model.CategoryDevice:
		selectVariables = append(selectVariables, "devices.name")
		joinDevice = true
	case model.CategoryShift:
		selectVariables = append(selectVariables, "TIME(products.created_at) >= '00:00:00' && TIME(products.created_at) <= '09:30:00'")
	case model.CategoryAttribute:
		if params.AttributeXAxis == nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("need AttributeXAxis when xAxis type is attribute"))
		}
		selectVariables = append(selectVariables, fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(products.`attribute`, '$.\"%v\"'))", *params.AttributeXAxis))
	}

	// group by
	if params.GroupBy != nil {
		selectQueries = append(selectQueries, "%v as group_by")
		groupColumns = append(groupColumns, "group_by")
		switch *params.GroupBy {
		case model.CategoryDate:
			selectVariables = append(selectVariables, "DATE(products.created_at)")
		case model.CategoryDevice:
			selectVariables = append(selectVariables, "devices.name")
			joinDevice = true
		case model.CategoryShift:
			selectVariables = append(selectVariables, "TIME(products.created_at) >= '00:00:00' && TIME(products.created_at) <= '09:30:00'")
		case model.CategoryAttribute:
			if params.AttributeGroup == nil {
				return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("need AttributeGroup when groupBy type is attribute"))
			}
			selectVariables = append(selectVariables, fmt.Sprintf("JSON_UNQUOTE(JSON_EXTRACT(products.`attribute`, '$.\"%v\"'))", *params.AttributeGroup))
		}
	}

	// join device
	if joinDevice {
		joins = append(joins, "JOIN devices ON products.device_id = devices.id")
	}

	// assemble joins
	query = query.Joins(strings.Join(joins, " "))

	// assemble selects
	query = query.Select(fmt.Sprintf(strings.Join(selectQueries, ", "), selectVariables...))

	// assemble groups
	query = query.Group(strings.Join(groupColumns, ", "))

	// time duration
	if len(params.Duration) > 0 {
		t := params.Duration[0]
		query = query.Where("products.created_at > ?", *t)
	}
	if len(params.Duration) > 1 {
		t := params.Duration[1]
		query = query.Where("products.created_at < ?", *t)
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
	}

	t2 := time.Now()
	fmt.Printf("query rows spand %v\n", t2.Sub(t1))
	results := scanRows(rows, params.GroupBy)
	t3 := time.Now()
	fmt.Printf("scan rows spand %v\n", t3.Sub(t2))

	if params.Limit != nil && *params.Limit < 0 {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("limit不能小于0"))
	}

	if params.YAxis != "Amount" {
		qualified := true
		if params.YAxis == "UnYield" {
			qualified = false
		}

		rows, err := query.Where("products.qualified = ?", qualified).Rows()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
		}
		qualifiedResults := scanRows(rows, params.GroupBy)

		return calYieldAnalysisResult(results, qualifiedResults, params.Limit, params.Sort)
	}

	eResult, err := calAmountAnalysisResult(results, params.Limit, params.Sort)
	if err != nil {
		return nil, err
	}
	return convertToEchartsResult(eResult), nil
}

func sortResult(result *echartsResult, isAsc bool) {
	var length = len(result.XAxisData)

	seriesData := result.SeriesData["data"]
	seriesAmountData := result.SeriesAmountData["data"]
	for i := 0; i < length-1; i++ {
		for j := 0; j < length-1-i; j++ {
			if isAsc && seriesData[j] < seriesData[j+1] {
				continue
			}

			if !isAsc && seriesData[j] > seriesData[j+1] {
				continue
			}

			s := seriesData[j]
			a := seriesAmountData[j]
			x := result.XAxisData[j]

			seriesData[j] = seriesData[j+1]
			seriesData[j+1] = s

			seriesAmountData[j] = seriesAmountData[j+1]
			seriesAmountData[j+1] = a

			result.XAxisData[j] = result.XAxisData[j+1]
			result.XAxisData[j+1] = x
		}
	}

	result.SeriesData["data"] = seriesData
	result.SeriesAmountData["data"] = seriesAmountData
}

func calYieldAnalysisResult(results, qualifiedResults []analysis, limit *int, sort *model.Sort) (*model.EchartsResult, error) {
	totalAmount, _ := calAmountAnalysisResult(results, nil, nil)
	yieldAmount, _ := calAmountAnalysisResult(qualifiedResults, nil, nil)

	t1 := time.Now()
	for i, item := range totalAmount.XAxisData {
		var index = findIndex(yieldAmount.XAxisData, item)
		for k, data := range totalAmount.SeriesData {
			yieldData := yieldAmount.SeriesData[k]
			total := data[i]

			if index < 0 || index >= len(yieldData) || total == 0 {
				data[i] = 0
			} else {
				yield := yieldData[index]
				value := yield / total
				data[i] = value
			}
			totalAmount.SeriesData[k] = data
		}
	}

	// sort
	if _, ok := totalAmount.SeriesData["data"]; ok && sort != nil {
		s := *sort
		sortResult(totalAmount, s.String() == "ASC")
	}

	// limit
	if limit != nil && *limit < len(totalAmount.XAxisData) {
		totalAmount.XAxisData = totalAmount.XAxisData[:*limit]
		for k, v := range totalAmount.SeriesData {
			totalAmount.SeriesData[k] = v[:*limit]
		}
		for k, v := range totalAmount.SeriesAmountData {
			totalAmount.SeriesAmountData[k] = v[:*limit]
		}
	}
	t2 := time.Now()
	fmt.Printf("[calYieldAnalysisResult] spend %v\n", t2.Sub(t1))

	return convertToEchartsResult(totalAmount), nil
}

func convertToEchartsResult(result *echartsResult) *model.EchartsResult {
	t1 := time.Now()
	seriesData := make(map[string]interface{})
	seriesAmountData := make(map[string]interface{})
	for k, v := range result.SeriesData {
		seriesData[k] = v
		seriesAmountData[k] = result.SeriesAmountData[k]
	}

	t2 := time.Now()
	fmt.Printf("[convertToEchartsResult] spend %v\n", t2.Sub(t1))
	return &model.EchartsResult{
		XAxisData:        result.XAxisData,
		SeriesData:       seriesData,
		SeriesAmountData: seriesAmountData,
	}
}

func findIndex(list []string, find string) int {
	for i, v := range list {
		if v == find {
			return i
		}
	}

	return -1
}

func calAmountAnalysisResult(scanResults []analysis, limit *int, sort *model.Sort) (*echartsResult, error) {
	t1 := time.Now()
	var xAxisMapData = make(map[string]int)
	var xAxisData []string
	var seriesMapData = make(map[string]map[string]float64)
	for i, result := range scanResults {
		xAxis := fmt.Sprint(result.Axis)
		if _, ok := xAxisMapData[xAxis]; !ok {
			xAxisMapData[xAxis] = i
			xAxisData = append(xAxisData, xAxis)
		}
		if seriesMap, ok := seriesMapData[fmt.Sprint(result.GroupBy)]; ok {
			seriesMap[fmt.Sprint(result.Axis)] = float64(result.Amount)
			seriesMapData[fmt.Sprint(result.GroupBy)] = seriesMap
		} else {
			seriesMap := map[string]float64{fmt.Sprint(result.Axis): float64(result.Amount)}
			seriesMapData[fmt.Sprint(result.GroupBy)] = seriesMap
		}
	}

	var seriesData = make(map[string][]float64)
	for _, item := range xAxisData {
		for k, seriesMap := range seriesMapData {
			sdv, ok := seriesData[k]
			var dataSet []float64
			if ok {
				dataSet = sdv
			} else {
				dataSet = make([]float64, 0)
			}

			var value float64
			if v, ok := seriesMap[item]; ok {
				value = v
			} else {
				value = 0
			}

			dataSet = append(dataSet, value)
			seriesData[k] = dataSet
		}
	}

	var seriesAmountData = make(map[string][]float64)
	for k, v := range seriesData {
		var data = append([]float64{}, v...)
		seriesAmountData[k] = data
	}

	var result = echartsResult{
		XAxisData:        xAxisData,
		SeriesData:       seriesData,
		SeriesAmountData: seriesAmountData,
	}

	// sort
	if _, ok := result.SeriesData["data"]; ok && sort != nil {
		s := *sort
		sortResult(&result, s.String() == "ASC")
	}

	// limit
	if limit != nil && *limit < len(result.XAxisData) {
		result.XAxisData = result.XAxisData[:*limit]
		for k, v := range result.SeriesData {
			result.SeriesData[k] = v[:*limit]
		}
		for k, v := range result.SeriesAmountData {
			result.SeriesAmountData[k] = v[:*limit]
		}
	}

	t2 := time.Now()
	fmt.Printf("[groupAnalyzeMaterial] spend %v\n", t2.Sub(t1))
	return &result, nil
}

func scanRows(rows *sql.Rows, groupBy *model.Category) []analysis {
	var results []analysis
	for rows.Next() {
		var result = analysis{GroupBy: "data"}
		var err error
		if groupBy != nil {
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
