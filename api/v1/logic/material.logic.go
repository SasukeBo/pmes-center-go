package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/model"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/jinzhu/copier"
	"strconv"
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
		//
		//ok, ng := countProductQualifiedForMaterial(&m)
		//out.Ok = ok
		//out.Ng = ng
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

	ok, ng := countProductQualifiedForMaterial(&material)
	out.Ok = ok
	out.Ng = ng

	return &out, nil
}

type qualifiedResult struct {
	Qualified bool
	Total     int64
}

func countProductQualifiedForMaterial(material *orm.Material) (int, int) {
	cacheKey := fmt.Sprintf("COUNT_PRODUCT_QUALIFIED_FOR_MATERIAL_%v", material.ID)
	if value, err := cache.GetString(cacheKey); err == nil {
		values := strings.Split(value, ",")
		if len(values) == 2 {
			ok, _ := strconv.Atoi(values[0])
			ng, _ := strconv.Atoi(values[1])
			cache.Set(cacheKey, value)
			return ok, ng
		}
	}

	currentVersion, err := material.GetCurrentVersion()
	if err != nil {
		return 0, 0
	}
	query := orm.Model(&orm.Product{}).Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where(
		"products.material_id = ? AND import_records.blocked = ? AND products.material_version_id = ?",
		material.ID, false, currentVersion.ID,
	)
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
	cache.Set(cacheKey, fmt.Sprintf("%v,%v", ok, ng))
	return ok, ng
}

func AnalyzeMaterial(ctx context.Context, id int, deviceID *int, versionID *int, duration []*time.Time) (*model.Material, error) {
	var material orm.Material
	if err := material.Get(uint(id)); err != nil {
		return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
	}
	var materialVersionID int
	if versionID == nil {
		version, err := material.GetCurrentVersion()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeActiveVersionNotFound, err)
		}
		materialVersionID = int(version.ID)
	} else {
		materialVersionID = *versionID
	}

	query := orm.Model(&orm.Product{}).Joins("JOIN import_records ON import_records.id = products.import_record_id")
	query = query.Where(
		"products.material_id = ? AND import_records.blocked = ? AND products.material_version_id = ?",
		material.ID, false, materialVersionID,
	)

	if len(duration) > 0 {
		query = query.Where("products.created_at > ?", *duration[0])
	}
	if len(duration) > 1 {
		query = query.Where("products.created_at < ?", *duration[1])
	}
	if deviceID != nil {
		query = query.Where("products.device_id = ?", *deviceID)
	}

	var ok int
	var ng int
	query.Where("products.qualified = ?", true).Count(&ok)
	query.Where("products.qualified = ?", false).Count(&ng)

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}

	out.Ok = ok
	out.Ng = ng
	return &out, nil
}

func MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	// SELECT
	query := orm.Model(&orm.Product{}).Select("materials.name AS name, COUNT(products.id) AS amount").Group("products.material_id")
	// JOIN
	query = query.Joins(`
	JOIN materials ON products.material_id = materials.id
	JOIN import_records ON import_records.id = products.import_record_id
	JOIN material_versions ON products.material_version_id = material_versions.id
	`)
	// WHERE
	query = query.Where("import_records.blocked = ? AND material_versions.active = ? ", false, true)

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

func GroupAnalyzeMaterial(ctx context.Context, analyzeInput model.GraphInput, versionID *int) (*model.EchartsResult, error) {
	return groupAnalyze(ctx, analyzeInput, "material", versionID)
}

func ProductAttributes(ctx context.Context, materialID int, versionID *int) ([]*model.ProductAttribute, error) {
	var version *orm.MaterialVersion
	var err error

	if versionID == nil {
		var material orm.Material
		if err := material.Get(uint(materialID)); err != nil {
			return nil, errormap.SendGQLError(ctx, err.GetCode(), err, "material")
		}

		version, err = material.GetCurrentVersion()
		if err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeActiveVersionNotFound, err, "material_version")
		}
	} else {
		var v orm.MaterialVersion
		if err := orm.Model(&orm.MaterialVersion{}).Where("id = ?", *versionID).Find(&v).Error; err != nil {
			return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_version")
		}
		version = &v
	}

	template, err := version.GetTemplate()
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "material_decode_template")
	}

	var outs []*model.ProductAttribute

	var rule orm.BarCodeRule
	if err := rule.Get(template.BarCodeRuleID); err != nil {
		log.Warnln(err)
		return outs, nil
	}

	itemsValue, ok := rule.Items["items"]
	if !ok {
		return outs, nil
	}

	items, ok := itemsValue.([]interface{})
	if !ok {
		return outs, nil
	}

	for _, v := range items {
		var out model.ProductAttribute
		value, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if iType, ok := value["type"]; ok {
			out.Type = fmt.Sprint(iType)
		}
		if token, ok := value["key"]; ok {
			out.Token = fmt.Sprint(token)
		}
		if label, ok := value["label"]; ok {
			out.Label = fmt.Sprint(label)
		}
		if name, ok := value["name"]; ok {
			out.Name = fmt.Sprint(name)
		}
		if len(out.Name) > 0 {
			out.Prefix = strings.ToUpper(out.Name[0:1])
		}
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
