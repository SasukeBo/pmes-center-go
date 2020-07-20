package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
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
	query := orm.Model(&orm.Product{}).Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("products.material_id = ? AND import_records.blocked = ?", id, false)
	query = query.Select("products.qualified, COUNT(products.id) as total")
	query = query.Group("products.qualified")
	rows, err := query.Rows()
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

func AnalyzeMaterial(ctx context.Context, searchInput model.Search) (*model.MaterialResult, error) {
	if searchInput.MaterialID == nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeRequestInputMissingFieldError, nil, "id")
	}
	var material orm.Material
	if err := material.Get(uint(*searchInput.MaterialID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	query := orm.Model(&orm.Product{}).Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("products.material_id = ? AND import_records.blocked = ?", material.ID, false)

	if searchInput.BeginTime != nil {
		query = query.Where("products.created_at > ?", *searchInput.BeginTime)
	}
	if searchInput.EndTime != nil {
		query = query.Where("products.created_at < ?", *searchInput.EndTime)
	}

	var ok int
	var ng int
	query.Where("products.qualified = ?", true).Count(&ok)
	query.Where("products.qualified = ?", false).Count(&ng)

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}
	return &model.MaterialResult{
		Material: &out,
		Ok:       ok,
		Ng:       ng,
	}, nil
}

func MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	query := orm.Model(&orm.Product{}).Select("materials.name AS name, COUNT(products.id) AS amount").Group("products.material_id")
	query = query.Joins("JOIN materials ON products.material_id = materials.id JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where("import_records.blocked = ?", false)

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

func GroupAnalyzeMaterial(ctx context.Context, analyzeInput model.GraphInput) (*model.EchartsResult, error) {
	return groupAnalyze(ctx, analyzeInput, "material")
}

func ProductAttributes(ctx context.Context, materialID int) ([]*model.ProductAttribute, error) {
	var template orm.DecodeTemplate
	if err := template.GetMaterialDefault(uint(materialID)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}

	var outs []*model.ProductAttribute
	for _, v := range template.ProductColumns {
		var out model.ProductAttribute
		value, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		out.Token = fmt.Sprint(value["Token"])
		out.Label = fmt.Sprint(value["Label"])
		out.Prefix = fmt.Sprint(value["Prefix"])
		outs = append(outs, &out)
	}

	return outs, nil
}

func LoadMaterial(ctx context.Context, materialID uint) *model.Material {
	var material orm.Material
	if err := material.Get(materialID); err != nil {
		return nil
	}
	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil
	}

	return &out
}
