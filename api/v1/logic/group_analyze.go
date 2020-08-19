package logic

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/gorm"
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

type queryProton struct {
	isRatio bool
	rows    *sql.Rows
}

func groupAnalyze(ctx context.Context, params model.GraphInput, target string, pVersionID *int) (*model.EchartsResult, error) {
	var out *model.EchartsResult
	var key string
	defer func() {
		if key != "" {
			_ = cache.Set(key, out)
		}
	}()
	for i, t := range params.Duration {
		nt := t.Truncate(time.Hour)
		params.Duration[i] = &nt
	}
	if content, err := json.Marshal(params); err == nil {
		key = fmt.Sprintf("%x-%s", md5.Sum(content), target)
		if v := cache.Get(key); v != nil {
			var ok bool
			out, ok = v.(*model.EchartsResult)
			if ok {
				return out, nil
			}
		}
	}

	if params.Limit != nil && *params.Limit < 0 {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeBadRequestParams, errors.New("limit不能小于0"))
	}
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

	// Version control and import_record block control
	versionID, blockIDs, err := getVersionIDAndBlockIDs(ctx, target, uint(params.TargetID), pVersionID)
	if err != nil {
		return nil, err
	}
	query = query.Where("products.material_version_id = ?", versionID)
	if len(blockIDs) > 0 {
		query = query.Where("products.import_record_id NOT IN (?)", blockIDs)
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

	var totalQueryCount = 1
	var rowsChan = make(chan queryProton, 2)
	go operateQuery(rowsChan, query, false)
	if params.YAxis != "Amount" {
		totalQueryCount++
		ratioQuery := query.Where("products.qualified = ?", params.YAxis == "Yield")
		go operateQuery(rowsChan, ratioQuery, true)
	}

	var receivedRowsCount int
	var ratioRows, rows *sql.Rows
	for {
		if receivedRowsCount >= totalQueryCount {
			break
		}
		select {
		case rowsProton := <-rowsChan:
			if rowsProton.isRatio {
				ratioRows = rowsProton.rows
			} else {
				rows = rowsProton.rows
			}
		}
		receivedRowsCount++
	}

	if rows == nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
	}
	results := scanRows(rows, params.GroupBy)

	if params.YAxis != "Amount" {
		if ratioRows == nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
		}
		qualifiedResults := scanRows(ratioRows, params.GroupBy)
		out, err = calYieldAnalysisResult(results, qualifiedResults, params.Limit, params.Sort)
		if err == nil && key != "" {
			_ = cache.Set(key, out)
		}
		return out, err
	}

	eResult, err := calAmountAnalysisResult(results, params.Limit, params.Sort)
	if err != nil {
		return nil, err
	}

	out = convertToEchartsResult(eResult)
	return out, nil
}

func operateQuery(rowsChan chan queryProton, query *gorm.DB, isRatio bool) {
	var result = queryProton{isRatio: isRatio}
	rows, err := query.Rows()
	if err != nil {
		rowsChan <- result
		return
	}
	result.rows = rows
	rowsChan <- result
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

	return convertToEchartsResult(totalAmount), nil
}

func convertToEchartsResult(result *echartsResult) *model.EchartsResult {
	seriesData := make(map[string]interface{})
	seriesAmountData := make(map[string]interface{})
	for k, v := range result.SeriesData {
		seriesData[k] = v
		seriesAmountData[k] = result.SeriesAmountData[k]
	}

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

func getVersionIDAndBlockIDs(ctx context.Context, target string, id uint, pVersionID *int) (int, []int, error) {
	var versionID int
	var materialID uint
	var blockIDs []int
	if pVersionID != nil {
		versionID = *pVersionID
	} else {
		var version orm.MaterialVersion
		if target == "material" {
			materialID = id
		} else if target == "device" {
			var device orm.Device
			if err := device.Get(id); err != nil {
				return versionID, blockIDs, errormap.SendGQLError(ctx, err.GetCode(), err, "device")
			}
			materialID = device.MaterialID
		}
		if err := version.GetActiveWithMaterialID(materialID); err != nil {
			return versionID, blockIDs, errormap.SendGQLError(ctx, err.GetCode(), err, "material_version")
		}
		versionID = int(version.ID)
	}

	err := orm.Model(&orm.ImportRecord{}).Where(
		"material_id = ? AND blocked = TRUE", materialID,
	).Pluck("id", &blockIDs).Error
	if err != nil {
		return versionID, blockIDs, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "import_record")
	}

	return versionID, blockIDs, nil
}
