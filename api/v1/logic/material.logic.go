package logic

import (
	"context"
	"crypto/md5"
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
	if value, err := cache.Get(cacheKey); err == nil {
		values := strings.Split(value, ",")
		if len(values) == 2 {
			ok, _ := strconv.Atoi(values[0])
			ng, _ := strconv.Atoi(values[1])
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
	for i, t := range duration {
		nt := t.Truncate(time.Hour)
		duration[i] = &nt
	}
	base := fmt.Sprintf("%s-%v-%v", "AnalyzeMaterial", id, duration)
	if deviceID != nil {
		base = base + fmt.Sprintf("-%v", *deviceID)
	}
	if versionID != nil {
		base = base + fmt.Sprintf("-%v", *versionID)
	}

	var out model.Material
	var key = fmt.Sprint(md5.Sum([]byte(base)))
	if err := cache.Scan(key, &out); err == nil {
		return &out, nil
	}

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

	var blockedIDs []int
	err := orm.Model(&orm.ImportRecord{}).Where(
		"material_id = ? AND blocked = true", material.ID,
	).Pluck("id", &blockedIDs).Error
	if err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "import_record")
	}

	query := orm.Model(&orm.Product{}).Select("count(id) AS total, qualified").Group("qualified")
	query = query.Where("material_version_id = ?", materialVersionID)
	if len(blockedIDs) > 0 {
		query = query.Where("import_record_id NOT IN (?)", blockedIDs)
	}

	if len(duration) > 0 {
		query = query.Where("products.created_at > ?", *duration[0])
	}
	if len(duration) > 1 {
		query = query.Where("products.created_at < ?", *duration[1])
	}
	if deviceID != nil {
		query = query.Where("products.device_id = ?", *deviceID)
	}

	var results []struct {
		Total     int  `json:"total"`
		Qualified bool `json:"qualified"`
	}
	if err := query.Scan(&results).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "products")
	}

	if err := copier.Copy(&out, &material); err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeTransferObjectError, err, "material")
	}
	for _, r := range results {
		if r.Qualified {
			out.Ok = r.Total
		} else {
			out.Ng = r.Total
		}
	}

	_ = cache.Set(key, &out)
	return &out, nil
}

type productGroupCount struct {
	name string
	ok   int
	ng   int
}

func MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	for i, t := range duration {
		nt := t.Truncate(time.Hour)
		duration[i] = &nt
	}
	var out model.EchartsResult
	var key = fmt.Sprint(md5.Sum([]byte(fmt.Sprintf("%s-%v-%v", "MaterialYieldTop", duration, limit))))
	if err := cache.Scan(key, &out); err == nil {
		return &out, nil
	}

	// SELECT
	query := orm.Model(&orm.Product{}).Select("COUNT(id) AS total, qualified").Group("qualified")

	if len(duration) > 0 {
		t := duration[0]
		query = query.Where("products.created_at > ?", *t)
	}

	if len(duration) > 1 {
		t := duration[1]
		query = query.Where("products.created_at < ?", *t)
	}

	t1 := time.Now()
	var materials []orm.Material
	if err := orm.Model(&orm.Material{}).Find(&materials).Error; err != nil {
		return nil, errormap.SendGQLError(ctx, errormap.ErrorCodeGetObjectFailed, err, "materials")
	}

	var productsChan = make(chan productGroupCount, 5)
	var groupCount int

	for _, material := range materials {
		versionID, blockIDs, err := getVersionIDAndBlockIDs(ctx, "material", material.ID, nil)
		if err != nil {
			continue
		}
		groupCount++
		var name = material.Name

		go func() {
			var result = productGroupCount{name: name}
			var scans []struct {
				Total     int
				Qualified bool
			}
			newQuery := query.Where(
				"material_version_id = ? ", versionID)
			if len(blockIDs) > 0 {
				newQuery = newQuery.Where("import_record_id NOT IN (?)", blockIDs)
			}
			if err := newQuery.Scan(&scans).Error; err != nil {
				productsChan <- result
				return
			}
			for _, scan := range scans {
				if scan.Qualified {
					result.ok = scan.Total
				} else {
					result.ng = scan.Total
				}
			}
			productsChan <- result
		}()
	}

	var results []productGroupCount
	var receivedCount int
	for {
		if receivedCount >= groupCount {
			break
		}
		select {
		case result := <-productsChan:
			results = append(results, result)
			fmt.Println(result)
		}
		receivedCount++
	}
	t2 := time.Now()
	log.Info("[MaterialYieldTop] query spend %v", t2.Sub(t1))

	var seriesData = make([]float64, 0)
	var xAxisData = make([]string, 0)

	for _, result := range results {
		xAxisData = append(xAxisData, result.name)
		var rate float64 = 0
		if total := result.ok + result.ng; total > 0 {
			rate = float64(result.ng) / float64(total)
		}
		seriesData = append(seriesData, rate)
	}

	// 排序
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

	out = model.EchartsResult{
		XAxisData: xAxisData[:limit],
		SeriesData: map[string]interface{}{
			"data": seriesData[:limit],
		},
	}
	_ = cache.Set(key, out)
	return &out, nil
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
