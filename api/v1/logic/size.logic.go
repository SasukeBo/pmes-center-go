package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/api/v1/model"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/orm"
)

// 尺寸不良率排行
func SizeUnYieldTop(ctx context.Context, groupInput model.GroupAnalyzeInput) (*model.EchartsResult, error) {
	query := orm.DB.Model(&orm.Product{}).Where("products.material_id = ?", groupInput.TargetID)

	// filters
	if groupInput.Filters != nil {
		for k, v := range groupInput.Filters {
			query = query.Where(fmt.Sprintf("JSON_EXTRACT(products.attribute, '$.\"%s\"') = ?", k), v)
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

	return &model.EchartsResult{
		XAxisData:        nil,
		SeriesData:       nil,
		SeriesAmountData: nil,
	}, nil

	type result struct {
		Name      string
		Qualified bool
	}
	var rs []result
	//for rows.Next() {
	//	var r result
	//	if err := rows.Scan(&r.Name, &r.Qualified); err != nil {
	//		continue
	//	}
	//	rs = append(rs, r)
	//}

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
