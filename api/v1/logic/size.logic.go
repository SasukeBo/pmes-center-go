package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	"strconv"
)

// 尺寸不良率排行
func SizeUnYieldTop(ctx context.Context, groupInput model.GraphInput) (*model.EchartsResult, error) {
	query := orm.DB.Model(&orm.Product{}).Where("products.material_id = ?", groupInput.TargetID)

	// filters
	if groupInput.Filters != nil {
		for k, v := range groupInput.Filters {
			switch k {
			case "device_id":
				query = query.Where("device_id = ?", v)
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
