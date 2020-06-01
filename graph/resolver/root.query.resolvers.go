package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	//"fmt"
	//"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/graph/logic"
	//"github.com/SasukeBo/ftpviewer/orm"
	//"github.com/google/uuid"
	//"github.com/jinzhu/gorm"
	//"math"
	//"strings"
	//"time"

	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/SasukeBo/ftpviewer/graph/model"
)

func (r *queryResolver) CurrentUser(ctx context.Context) (*model.User, error) {
	return logic.CurrentUser(ctx)
}

func (r *queryResolver) Products(ctx context.Context, searchInput model.Search, page *int, limit int, offset *int) (*model.ProductWrap, error) {
	return nil, nil
	//if searchInput.MaterialID == nil {
	//	return nil, NewGQLError("料号ID不能为空", "searchInput.MaterialID is nil")
	//}
	//oset := 0
	//if offset != nil {
	//	oset = *offset
	//} else if page != nil {
	//	if *page < 1 {
	//		return nil, NewGQLError("页数不能小于1", "")
	//	}
	//	oset = (*page - 1) * limit
	//}
	//
	//var conditions []string
	//var vars []interface{}
	//material := orm.GetMaterialWithID(*searchInput.MaterialID)
	//if material == nil {
	//	return nil, NewGQLError("您所查找的料号不存在", fmt.Sprintf("get material with id = %v failed", *searchInput.MaterialID))
	//}
	//
	//end := searchInput.EndTime
	//if end == nil {
	//	t := time.Now()
	//	end = &t
	//}
	//begin := searchInput.BeginTime
	//if begin == nil {
	//	t := end.AddDate(-1, 0, 0)
	//	begin = &t
	//}
	//
	//fileIDs, err := logic.NeedFetch(material, begin, end)
	//if err != nil {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(false), Message: stringP(err.Error())}
	//	return &model.ProductWrap{Status: status}, nil
	//}
	//
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
	//	return &model.ProductWrap{Status: status}, nil
	//}
	//
	//conditions = append(conditions, "material_id = ?")
	//vars = append(vars, material.ID)
	//if searchInput.DeviceID != nil {
	//	device := orm.GetDeviceWithID(*searchInput.DeviceID)
	//	if device != nil {
	//		conditions = append(conditions, "device_id = ?")
	//		vars = append(vars, device.ID)
	//	}
	//}
	//conditions = append(conditions, "created_at < ?")
	//vars = append(vars, end)
	//conditions = append(conditions, "created_at > ?")
	//vars = append(vars, begin)
	//
	//if lineID, ok := searchInput.Extra["lineID"]; ok {
	//	conditions = append(conditions, "line_id = ?")
	//	vars = append(vars, lineID)
	//}
	//
	//if mouldID, ok := searchInput.Extra["mouldID"]; ok {
	//	conditions = append(conditions, "mould_id = ?")
	//	vars = append(vars, mouldID)
	//}
	//
	//if jigID, ok := searchInput.Extra["jigID"]; ok {
	//	conditions = append(conditions, "jig_id = ?")
	//	vars = append(vars, jigID)
	//}
	//
	//if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
	//	conditions = append(conditions, "shift_number = ?")
	//	vars = append(vars, shiftNumber)
	//}
	//
	//fmt.Println(conditions)
	//cond := strings.Join(conditions, " AND ")
	//var products []orm.Product
	//if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("id asc").Offset(oset).Limit(limit).Find(&products).Error; err != nil {
	//	if err == gorm.ErrRecordNotFound { // 无数据
	//		return &model.ProductWrap{
	//			TableHeader: nil,
	//			Products:    nil,
	//			Status:      nil,
	//			Total:       intP(0),
	//		}, nil
	//	}
	//
	//	return nil, NewGQLError("获取数据失败，请重试", err.Error())
	//}
	//
	//var total int
	//if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Count(&total).Error; err != nil {
	//	return nil, NewGQLError("统计产品数量失败", err.Error())
	//}
	//
	//var productUUIDs []string
	//for _, p := range products {
	//	productUUIDs = append(productUUIDs, p.UUID)
	//}
	//
	//rows, err := orm.DB.Raw(`
	//SELECT pv.product_uuid, p.name, pv.v FROM point_values AS pv
	//JOIN points AS p ON pv.point_id = p.id
	//WHERE pv.product_uuid IN (?)
	//ORDER BY pv.product_uuid, p.index
	//`, productUUIDs).Rows()
	//if err != nil {
	//	return nil, NewGQLError("获取产品尺寸数据失败", err.Error())
	//}
	//defer rows.Close()
	//
	//var uuid, name string
	//var value float64
	//productPointValueMap := make(map[string]map[string]interface{})
	//for rows.Next() {
	//	rows.Scan(&uuid, &name, &value)
	//	if p, ok := productPointValueMap[uuid]; ok {
	//		p[name] = value
	//		continue
	//	}
	//
	//	productPointValueMap[uuid] = map[string]interface{}{name: value}
	//}
	//
	//var outProducts []*model.Product
	//for _, i := range products {
	//	p := i
	//	op := &model.Product{
	//		ID:          &p.ID,
	//		UUID:        &p.UUID,
	//		MaterialID:  &p.MaterialID,
	//		DeviceID:    &p.DeviceID,
	//		Qualified:   &p.Qualified,
	//		CreatedAt:   &p.CreatedAt,
	//		D2code:      &p.D2Code,
	//		LineID:      &p.LineID,
	//		JigID:       &p.JigID,
	//		MouldID:     &p.MouldID,
	//		ShiftNumber: &p.ShiftNumber,
	//	}
	//	if mp, ok := productPointValueMap[p.UUID]; ok {
	//		op.PointValue = mp
	//	}
	//	outProducts = append(outProducts, op)
	//}
	//
	//var sizeIDs []int
	//orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs)
	//
	//var pointNames []string
	//orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index asc").Pluck("name", &pointNames)
	//
	//status := model.FetchStatus{Pending: boolP(false)}
	//return &model.ProductWrap{
	//	TableHeader: pointNames,
	//	Products:    outProducts,
	//	Status:      &status,
	//	Total:       &total,
	//}, nil
}

func (r *queryResolver) ExportProducts(ctx context.Context, searchInput model.Search) (string, error) {
	return "", nil
	//if searchInput.MaterialID == nil {
	//	return "error", NewGQLError("料号ID不能为空")
	//}
	//material := orm.GetMaterialWithID(*searchInput.MaterialID)
	//if material == nil {
	//	return "error", NewGQLError("料号不存在")
	//}
	//
	//// 拼接查询条件
	//var conditions []string
	//var vars []interface{}
	//conditions = append(conditions, "material_id = ?")
	//vars = append(vars, material.ID)
	//
	//if searchInput.EndTime == nil {
	//	t := time.Now()
	//	searchInput.EndTime = &t
	//}
	//
	//if searchInput.BeginTime == nil {
	//	t := searchInput.EndTime.AddDate(-1, 0, 0)
	//	searchInput.BeginTime = &t
	//}
	//conditions = append(conditions, "created_at < ?")
	//vars = append(vars, searchInput.EndTime)
	//conditions = append(conditions, "created_at > ?")
	//vars = append(vars, searchInput.BeginTime)
	//
	//if searchInput.DeviceID != nil {
	//	device := orm.GetDeviceWithID(*searchInput.DeviceID)
	//	if device != nil {
	//		conditions = append(conditions, "device_id = ?")
	//		vars = append(vars, device.ID)
	//	}
	//}
	//
	//if lineID, ok := searchInput.Extra["lineID"]; ok {
	//	conditions = append(conditions, "line_id = ?")
	//	vars = append(vars, lineID)
	//}
	//
	//if mouldID, ok := searchInput.Extra["mouldID"]; ok {
	//	conditions = append(conditions, "mould_id = ?")
	//	vars = append(vars, mouldID)
	//}
	//
	//if jigID, ok := searchInput.Extra["jigID"]; ok {
	//	conditions = append(conditions, "jig_id = ?")
	//	vars = append(vars, jigID)
	//}
	//
	//if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
	//	conditions = append(conditions, "shift_number = ?")
	//	vars = append(vars, shiftNumber)
	//}
	//
	//opID := uuid.New().String()
	//condition := strings.Join(conditions, " AND ")
	//go logic.HandleExport(opID, material, searchInput, condition, vars...)
	//
	//return opID, nil
}

func (r *queryResolver) AnalyzePoint(ctx context.Context, searchInput model.Search, limit int, offset int, pattern *string) (*model.PointResultsWrap, error) {
	return nil, nil
	//if searchInput.MaterialID == nil {
	//	return nil, NewGQLError("料号ID不能为空", "")
	//}
	//
	//material := orm.GetMaterialWithID(*searchInput.MaterialID)
	//if material == nil {
	//	return nil, NewGQLError("没有找到该尺寸所属的料号", "")
	//}
	//
	//var sizeIDs []int
	//if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
	//	return nil, NewGQLError("获取该料号的尺寸信息发生错误", err.Error())
	//}
	//
	//pointSql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)
	//if pattern != nil {
	//	pointSql = pointSql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *pattern))
	//}
	//
	//var points []orm.Point
	//if err := pointSql.Order("points.index ASC").Limit(limit).Offset(offset).Find(&points).Error; err != nil {
	//	return nil, NewGQLError("获取尺寸点位失败", err.Error())
	//}
	//
	//var total int
	//if err := pointSql.Count(&total).Error; err != nil {
	//	return nil, NewGQLError("统计检测点位总数时发生错误", err.Error())
	//}
	//
	//end := searchInput.EndTime
	//if end == nil {
	//	t := time.Now()
	//	end = &t
	//}
	//begin := searchInput.BeginTime
	//if begin == nil {
	//	t := end.AddDate(0, -1, 0)
	//	begin = &t
	//}
	//
	//var conds []string
	//var vars []interface{}
	//
	//if searchInput.DeviceID != nil {
	//	conds = append(conds, "device_id = ?")
	//	vars = append(vars, *searchInput.DeviceID)
	//}
	//conds = append(conds, "created_at > ?")
	//vars = append(vars, *begin)
	//conds = append(conds, "created_at < ?")
	//vars = append(vars, *end)
	//
	//if lineID, ok := searchInput.Extra["lineID"]; ok {
	//	conds = append(conds, "line_id = ?")
	//	vars = append(vars, lineID)
	//}
	//
	//if mouldID, ok := searchInput.Extra["mouldID"]; ok {
	//	conds = append(conds, "mould_id = ?")
	//	vars = append(vars, mouldID)
	//}
	//
	//if jigID, ok := searchInput.Extra["jigID"]; ok {
	//	conds = append(conds, "jig_id = ?")
	//	vars = append(vars, jigID)
	//}
	//
	//if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
	//	conds = append(conds, "shift_number = ?")
	//	vars = append(vars, shiftNumber)
	//}
	//
	//cond := strings.Join(conds, " AND ")
	//var pointResults []*model.PointResult
	//mpvs := make(map[int][]float64)
	//sql := orm.DB.Model(&orm.PointValue{}).Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Select("point_values.v")
	//
	//for _, p := range points {
	//	var pointValues []float64
	//	rows, err := sql.Where("point_values.point_id = ?", p.ID).Rows()
	//	if err != nil {
	//		rows.Close()
	//		mpvs[p.ID] = pointValues
	//		continue
	//	}
	//
	//	for rows.Next() {
	//		var v float64
	//		rows.Scan(&v)
	//		pointValues = append(pointValues, v)
	//	}
	//	rows.Close()
	//	mpvs[p.ID] = pointValues
	//}
	//
	//for _, p := range points {
	//	data := mpvs[p.ID]
	//	s, cp, cpk, avg, ok, total, valueSet := logic.AnalyzePointValues(p, data)
	//	min, max, values, freqs, distribution := logic.Distribute(s, avg, valueSet)
	//
	//	point := p
	//	pointResult := &model.PointResult{
	//		Total: &total,
	//		S:     &s,
	//		Ok:    &ok,
	//		Ng:    intP(total - ok),
	//		Cp:    &cp,
	//		Cpk:   &cpk,
	//		Avg:   &avg,
	//		Max:   &max,
	//		Min:   &min,
	//		Dataset: map[string]interface{}{
	//			"values":       values,
	//			"freqs":        freqs,
	//			"distribution": distribution,
	//		},
	//		Point: &model.Point{
	//			ID:         &point.ID,
	//			Name:       &point.Name,
	//			UpperLimit: &point.UpperLimit,
	//			Nominal:    &point.Nominal,
	//			LowerLimit: &point.LowerLimit,
	//		},
	//	}
	//
	//	pointResults = append(pointResults, pointResult)
	//}
	//
	//return &model.PointResultsWrap{
	//	PointResults: pointResults,
	//	Total:        total,
	//}, nil
}

func (r *queryResolver) TotalPointYield(ctx context.Context, searchInput model.Search, pattern *string) ([]*model.YieldWrap, error) {
	return nil, nil
	//if searchInput.MaterialID == nil {
	//	return nil, NewGQLError("料号ID不能为空", "")
	//}
	//
	//material := orm.GetMaterialWithID(*searchInput.MaterialID)
	//if material == nil {
	//	return nil, NewGQLError("没有找到该尺寸所属的料号", "")
	//}
	//
	//var sizeIDs []int
	//if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
	//	return nil, NewGQLError("获取该料号的尺寸信息发生错误", err.Error())
	//}
	//
	//pointSql := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs)
	//if pattern != nil {
	//	pointSql = pointSql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *pattern))
	//}
	//
	//var points []orm.Point
	//if err := pointSql.Order("points.index ASC").Find(&points).Error; err != nil {
	//	return nil, NewGQLError("获取尺寸点位失败", err.Error())
	//}
	//
	//end := searchInput.EndTime
	//if end == nil {
	//	t := time.Now()
	//	end = &t
	//}
	//begin := searchInput.BeginTime
	//if begin == nil {
	//	t := end.AddDate(0, -1, 0)
	//	begin = &t
	//}
	//
	//var conds []string
	//var vars []interface{}
	//
	//if searchInput.DeviceID != nil {
	//	conds = append(conds, "device_id = ?")
	//	vars = append(vars, *searchInput.DeviceID)
	//}
	//conds = append(conds, "created_at > ?")
	//vars = append(vars, *begin)
	//conds = append(conds, "created_at < ?")
	//vars = append(vars, *end)
	//
	//if lineID, ok := searchInput.Extra["lineID"]; ok {
	//	conds = append(conds, "line_id = ?")
	//	vars = append(vars, lineID)
	//}
	//
	//if mouldID, ok := searchInput.Extra["mouldID"]; ok {
	//	conds = append(conds, "mould_id = ?")
	//	vars = append(vars, mouldID)
	//}
	//
	//if jigID, ok := searchInput.Extra["jigID"]; ok {
	//	conds = append(conds, "jig_id = ?")
	//	vars = append(vars, jigID)
	//}
	//
	//if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
	//	conds = append(conds, "shift_number = ?")
	//	vars = append(vars, shiftNumber)
	//}
	//
	//cond := strings.Join(conds, " AND ")
	//mpvs := make(map[int][]float64)
	//sql := orm.DB.Model(&orm.PointValue{}).Joins("LEFT JOIN products ON point_values.product_uuid = products.uuid").Where(cond, vars...).Select("point_values.v")
	//
	//for _, p := range points {
	//	var pointValues []float64
	//	rows, err := sql.Where("point_values.point_id = ?", p.ID).Rows()
	//	if err != nil {
	//		rows.Close()
	//		mpvs[p.ID] = pointValues
	//		continue
	//	}
	//
	//	for rows.Next() {
	//		var v float64
	//		rows.Scan(&v)
	//		pointValues = append(pointValues, v)
	//	}
	//	rows.Close()
	//	mpvs[p.ID] = pointValues
	//}
	//
	//var out []*model.YieldWrap
	//for _, p := range points {
	//	data := mpvs[p.ID]
	//	total := len(data)
	//	ok := 0
	//	for _, v := range data {
	//		if p.NotValid(v) {
	//			continue
	//		}
	//
	//		if v >= p.LowerLimit && v <= p.UpperLimit {
	//			ok++
	//		}
	//	}
	//	o := &model.YieldWrap{
	//		Name:  p.Name,
	//		Value: float64(ok) / float64(total),
	//	}
	//	out = append(out, o)
	//}
	//
	//return out, nil
}

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, searchInput model.Search) (*model.MaterialResult, error) {
	return nil, nil
	//if searchInput.MaterialID == nil {
	//	return nil, NewGQLError("料号ID不能为空", "searchInput.ID can't be empty")
	//}
	//material := orm.GetMaterialWithID(*searchInput.MaterialID)
	//if material == nil {
	//	return nil, NewGQLError("料号不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.MaterialID))
	//}
	//beginTime := searchInput.BeginTime
	//endTime := searchInput.EndTime
	//if endTime == nil {
	//	t := time.Now()
	//	endTime = &t
	//}
	//if beginTime == nil {
	//	t := endTime.AddDate(-1, 0, 0)
	//	beginTime = &t
	//}
	//
	//fileIDs, err := logic.NeedFetch(material, beginTime, endTime)
	//if err != nil {
	//	return nil, err
	//}
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
	//	return &model.MaterialResult{Status: status}, nil
	//}
	//
	//conditions := []string{"material_id = ?", "created_at < ?", "created_at > ?"}
	//vars := []interface{}{searchInput.MaterialID, endTime, beginTime}
	//
	//if lineID, ok := searchInput.Extra["lineID"]; ok {
	//	conditions = append(conditions, "line_id = ?")
	//	vars = append(vars, lineID)
	//}
	//
	//if mouldID, ok := searchInput.Extra["mouldID"]; ok {
	//	conditions = append(conditions, "mould_id = ?")
	//	vars = append(vars, mouldID)
	//}
	//
	//if jigID, ok := searchInput.Extra["jigID"]; ok {
	//	conditions = append(conditions, "jig_id = ?")
	//	vars = append(vars, jigID)
	//}
	//
	//if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
	//	conditions = append(conditions, "shift_number = ?")
	//	vars = append(vars, shiftNumber)
	//}
	//conditions = append(conditions, "qualified = ?")
	//
	//var ok int
	//var ng int
	//cond := strings.Join(conditions, " AND ")
	//varsQualified := append(vars, 1)
	//orm.DB.Model(&orm.Product{}).Where(cond, varsQualified...).Count(&ok)
	//varsUnqualified := append(vars, 0)
	//orm.DB.Model(&orm.Product{}).Where(cond, varsUnqualified...).Count(&ng)
	//out := model.Material{
	//	ID:            material.ID,
	//	Name:          material.Name,
	//	CustomerCode:  stringP(material.CustomerCode),
	//	ProjectRemark: stringP(material.ProjectRemark),
	//}
	//
	//return &model.MaterialResult{
	//	Material: &out,
	//	Ok:       &ok,
	//	Ng:       &ng,
	//	Status:   &model.FetchStatus{Pending: boolP(false)},
	//}, nil
}

func (r *queryResolver) Materials(ctx context.Context, page int, limit int) (*model.MaterialWrap, error) {
	return nil, nil
	//var materials []orm.Material
	//if page < 1 {
	//	return nil, NewGQLError("页数不能小于1", "page < 1")
	//}
	//offset := (page - 1) * limit
	//if err := orm.DB.Order("id desc").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
	//	return nil, NewGQLError("获取料号信息失败", err.Error())
	//}
	//var outs []*model.Material
	//for _, i := range materials {
	//	v := i
	//	outs = append(outs, &model.Material{
	//		ID:            v.ID,
	//		Name:          v.Name,
	//		CustomerCode:  stringP(v.CustomerCode),
	//		ProjectRemark: stringP(v.ProjectRemark),
	//	})
	//}
	//var count int
	//if err := orm.DB.Model(&orm.Material{}).Count(&count).Error; err != nil {
	//	return nil, NewGQLError("统计料号数量失败", err.Error())
	//}
	//return &model.MaterialWrap{
	//	Total:     &count,
	//	Materials: outs,
	//}, nil
}

func (r *queryResolver) MaterialsWithSearch(ctx context.Context, offset int, limit int, search *string) (*model.MaterialWrap, error) {
	return nil, nil
	//var materials []orm.Material
	//db := orm.DB.Order("id desc").Limit(limit).Offset(offset)
	//
	//if search != nil {
	//	pattern := "%" + *search + "%"
	//	db = db.Where("name LIKE ? OR customer_code LIKE ? OR project_remark LIKE ?", pattern, pattern, pattern)
	//}
	//if err := db.Find(&materials).Error; err != nil {
	//	return nil, NewGQLError("获取料号信息失败", err.Error())
	//}
	//var outs []*model.Material
	//for _, i := range materials {
	//	v := i
	//	outs = append(outs, &model.Material{
	//		ID:            v.ID,
	//		Name:          v.Name,
	//		CustomerCode:  stringP(v.CustomerCode),
	//		ProjectRemark: stringP(v.ProjectRemark),
	//	})
	//}
	//var count int
	//if err := orm.DB.Model(&orm.Material{}).Count(&count).Error; err != nil {
	//	return nil, NewGQLError("统计料号数量失败", err.Error())
	//}
	//return &model.MaterialWrap{
	//	Total:     &count,
	//	Materials: outs,
	//}, nil
}

func (r *queryResolver) AnalyzeDevice(ctx context.Context, searchInput model.Search) (*model.DeviceResult, error) {
	return nil, nil
	//if searchInput.DeviceID == nil {
	//	return nil, NewGQLError("设备ID不能为空", "searchInput.DeviceID can't be empty")
	//}
	//device := orm.GetDeviceWithID(*searchInput.DeviceID)
	//if device == nil {
	//	return nil, NewGQLError("设备不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.DeviceID))
	//}
	//material := orm.GetMaterialWithID(device.MaterialID)
	//if material == nil {
	//	return nil, NewGQLError("设备生产的料号不存在", fmt.Sprintf("get material with id = %v failed", device.MaterialID))
	//}
	//
	//beginTime := searchInput.BeginTime
	//endTime := searchInput.EndTime
	//if endTime == nil {
	//	t := time.Now()
	//	endTime = &t
	//}
	//if beginTime == nil {
	//	t := endTime.AddDate(-1, 0, 0)
	//	beginTime = &t
	//}
	//
	//out := model.Device{
	//	ID:   &device.ID,
	//	Name: &device.Name,
	//}
	//fileIDs, err := logic.NeedFetch(material, beginTime, endTime)
	//if err != nil {
	//	return nil, err
	//}
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内设备数据")}
	//	return &model.DeviceResult{Status: status, Device: &out}, nil
	//}
	//
	//var ok int
	//var ng int
	//orm.DB.Model(&orm.Product{}).Where(
	//	"device_id = ? and created_at < ? and created_at > ? and qualified = 1",
	//	searchInput.DeviceID, endTime, beginTime,
	//).Count(&ok)
	//orm.DB.Model(&orm.Product{}).Where(
	//	"device_id = ? and created_at < ? and created_at > ? and qualified = 0",
	//	searchInput.DeviceID, endTime, beginTime,
	//).Count(&ng)
	//
	//return &model.DeviceResult{
	//	Device: &out,
	//	Ok:     &ok,
	//	Ng:     &ng,
	//}, nil
}

func (r *queryResolver) Devices(ctx context.Context, materialID int) ([]*model.Device, error) {
	return nil, nil
	//var devices []orm.Device
	//if err := orm.DB.Where("material_id = ?", materialID).Find(&devices).Error; err != nil {
	//	return nil, NewGQLError("获取设备信息失败", err.Error())
	//}
	//var outs []*model.Device
	//for _, i := range devices {
	//	v := i
	//	outs = append(outs, &model.Device{
	//		ID:   &v.ID,
	//		Name: &v.Name,
	//	})
	//}
	//return outs, nil
}

func (r *queryResolver) Sizes(ctx context.Context, page int, limit int, materialID int) (*model.SizeWrap, error) {
	return nil, nil
	//var sizes []orm.Size
	//if page < 1 {
	//	return nil, NewGQLError("页数不能小于1", "page < 1")
	//}
	//offset := (page - 1) * limit
	//if err := orm.DB.Where("material_id = ?", materialID).Limit(limit).Offset(offset).Find(&sizes).Error; err != nil {
	//	return nil, NewGQLError("获取尺寸信息失败", err.Error())
	//}
	//var outs []*model.Size
	//for _, v := range sizes {
	//	s := v
	//	outs = append(outs, &model.Size{
	//		ID:         &s.ID,
	//		Name:       &s.Name,
	//		MaterialID: &s.MaterialID,
	//	})
	//}
	//var count int
	//if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", materialID).Count(&count).Error; err != nil {
	//	return nil, NewGQLError("统计尺寸数量失败", err.Error())
	//}
	//
	//return &model.SizeWrap{
	//	Total: &count,
	//	Sizes: outs,
	//}, nil
}

func (r *queryResolver) DataFetchFinishPercent(ctx context.Context, fileIDs []*int) (float64, error) {
	return 0, nil
	//total := len(fileIDs)
	//if total == 0 {
	//	return 0, nil
	//}
	//
	//result := struct {
	//	Finished int
	//	Total    int
	//}{}
	//err := orm.DB.Table("files").Where("id in (?)", fileIDs).Select("SUM(total_rows) as total, SUM(finished_rows) as finished").First(&result).Error
	//if err != nil {
	//	return 0, NewGQLError("查询数据导入完成度失败", err.Error())
	//}
	//
	//if result.Total == 0 {
	//	return 0, nil
	//}
	//percent := float64(result.Finished) / float64(result.Total)
	//
	//return math.Round(percent*100) / 100, nil
}

func (r *queryResolver) ExportFinishPercent(ctx context.Context, opID string) (*model.ExportResponse, error) {
	return nil, nil
	//rsp, err := logic.CheckExport(opID)
	//if err != nil {
	//	return nil, NewGQLError(rsp.Message, err.Error())
	//}
	//
	//return rsp, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
