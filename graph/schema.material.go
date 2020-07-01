package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"math"
	"strings"
	"time"
)

func (r *mutationResolver) UpdateMaterial(ctx context.Context, input model.MaterialUpdateInput) (*model.Material, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	var material orm.Material
	if err := orm.DB.Model(&orm.Material{}).Where("id = ?", input.ID).First(&material).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewGQLError("料号不存在", err.Error())
		}

		return nil, NewGQLError("获取料号失败", err.Error())
	}

	if input.ProjectRemark != nil {
		material.ProjectRemark = *input.ProjectRemark
	}

	if input.CustomerCode != nil {
		material.CustomerCode = *input.CustomerCode
	}

	if err := orm.DB.Save(&material).Error; err != nil {
		return nil, NewGQLError("保存料号失败", err.Error())
	}

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, NewGQLError("数据转换发生错误", err.Error())
	}

	return &out, nil
}

func (r *mutationResolver) AddMaterial(ctx context.Context, input model.MaterialCreateInput) (*model.Material, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return nil, err
	}

	if input.Name == "" {
		return nil, NewGQLError("厂内料号不能为空", "empty material code")
	}

	if !logic.IsMaterialExist(input.Name) {
		return nil, NewGQLError("FTP服务器现在没有该料号的数据。", "IsMaterialExist false")
	}

	material := orm.GetMaterialWithName(input.Name)
	if material != nil {
		return nil, NewGQLError("料号已经存在，请确认你的输入。", "find material, can't create another one.")
	}

	m := orm.Material{Name: input.Name}
	if input.CustomerCode != nil {
		m.CustomerCode = *input.CustomerCode
	}
	if input.ProjectRemark != nil {
		m.ProjectRemark = *input.ProjectRemark
	}
	if err := orm.DB.Create(&m).Error; err != nil {
		return nil, NewGQLError("创建料号失败", err.Error())
	}

	end := time.Now()
	// 默认取近一年的数据
	begin := end.AddDate(-1, 0, 0)
	logic.FetchData(&m, &begin, &end)

	var out model.Material
	if err := copier.Copy(&out, &m); err != nil {
		return nil, NewGQLError("转换数据发生错误", err.Error())
	}

	return &out, nil
}

func (r *queryResolver) DataFetchFinishPercent(ctx context.Context, fileIDs []*int) (float64, error) {
	total := len(fileIDs)
	if total == 0 {
		return 0, nil
	}

	result := struct {
		Finished int
		Total    int
	}{}
	err := orm.DB.Table("files").Where("id in (?)", fileIDs).Select("SUM(total_rows) as total, SUM(finished_rows) as finished").First(&result).Error
	if err != nil {
		return 0, NewGQLError("查询数据导入完成度失败", err.Error())
	}

	if result.Total == 0 {
		return 0, nil
	}
	percent := float64(result.Finished) / float64(result.Total)

	return math.Round(percent*100) / 100, nil
}

func (r *queryResolver) AnalyzeMaterial(ctx context.Context, searchInput model.Search) (*model.MaterialResult, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("料号ID不能为空", "searchInput.ID can't be empty")
	}
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("料号不存在", fmt.Sprintf("get device with id = %v failed", *searchInput.MaterialID))
	}
	beginTime := searchInput.BeginTime
	endTime := searchInput.EndTime
	if endTime == nil {
		t := time.Now()
		endTime = &t
	}
	if beginTime == nil {
		t := endTime.AddDate(-1, 0, 0)
		beginTime = &t
	}

	// TODO: 关闭自动拉取数据
	//fileIDs, err := logic.FetchData(material, beginTime, endTime)
	//if err != nil {
	//	return nil, err
	//}
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
	//	return &model.MaterialResult{Status: status}, nil
	//}

	conditions := []string{"material_id = ?", "created_at < ?", "created_at > ?"}
	vars := []interface{}{searchInput.MaterialID, endTime, beginTime}

	if lineID, ok := searchInput.Extra["lineID"]; ok {
		conditions = append(conditions, "line_id = ?")
		vars = append(vars, lineID)
	}

	if mouldID, ok := searchInput.Extra["mouldID"]; ok {
		conditions = append(conditions, "mould_id = ?")
		vars = append(vars, mouldID)
	}

	if jigID, ok := searchInput.Extra["jigID"]; ok {
		conditions = append(conditions, "jig_id = ?")
		vars = append(vars, jigID)
	}

	if shiftNumber, ok := searchInput.Extra["shiftNumber"]; ok {
		conditions = append(conditions, "shift_number = ?")
		vars = append(vars, shiftNumber)
	}
	conditions = append(conditions, "qualified = ?")

	var ok int
	var ng int
	cond := strings.Join(conditions, " AND ")
	varsQualified := append(vars, 1)
	orm.DB.Model(&orm.Product{}).Where(cond, varsQualified...).Count(&ok)
	varsUnqualified := append(vars, 0)
	orm.DB.Model(&orm.Product{}).Where(cond, varsUnqualified...).Count(&ng)

	var out model.Material
	if err := copier.Copy(&out, &material); err != nil {
		return nil, NewGQLError("数据转换发生错误", err.Error())
	}
	return &model.MaterialResult{
		Material: &out,
		Ok:       ok,
		Ng:       ng,
	}, nil
}

func (r *queryResolver) Materials(ctx context.Context, page int, limit int) (*model.MaterialWrap, error) {
	var materials []orm.Material
	if page < 1 {
		return nil, NewGQLError("页数不能小于1", "page < 1")
	}
	offset := (page - 1) * limit
	if err := orm.DB.Order("id desc").Limit(limit).Offset(offset).Find(&materials).Error; err != nil {
		return nil, NewGQLError("获取料号信息失败", err.Error())
	}
	var outs []*model.Material
	for _, i := range materials {
		v := i
		outs = append(outs, &model.Material{
			ID:            v.ID,
			Name:          v.Name,
			CustomerCode:  v.CustomerCode,
			ProjectRemark: v.ProjectRemark,
		})
	}
	var count int
	if err := orm.DB.Model(&orm.Material{}).Count(&count).Error; err != nil {
		return nil, NewGQLError("统计料号数量失败", err.Error())
	}
	return &model.MaterialWrap{
		Total:     count,
		Materials: outs,
	}, nil
}

func (r *queryResolver) MaterialsWithSearch(ctx context.Context, offset int, limit int, search *string) (*model.MaterialWrap, error) {
	var materials []orm.Material
	db := orm.DB.Order("id desc").Limit(limit).Offset(offset)

	if search != nil {
		pattern := "%" + *search + "%"
		db = db.Where("name LIKE ? OR customer_code LIKE ? OR project_remark LIKE ?", pattern, pattern, pattern)
	}
	if err := db.Find(&materials).Error; err != nil {
		return nil, NewGQLError("获取料号信息失败", err.Error())
	}
	var outs []*model.Material
	for _, m := range materials {
		var out model.Material
		if err := copier.Copy(&out, &m); err != nil {
			continue
		}
		outs = append(outs, &out)
	}
	var count int
	if err := orm.DB.Model(&orm.Material{}).Count(&count).Error; err != nil {
		return nil, NewGQLError("统计料号数量失败", err.Error())
	}
	return &model.MaterialWrap{
		Total:     count,
		Materials: outs,
	}, nil
}

func (r *mutationResolver) DeleteMaterial(ctx context.Context, id int) (string, error) {
	if err := logic.Authenticate(ctx); err != nil {
		return "error", err
	}

	tx := orm.DB.Begin()
	defer tx.Rollback()

	var sizeIDs []int
	if err := tx.Model(&orm.Size{}).Where("material_id = ?", id).Pluck("id", &sizeIDs).Error; err != nil {
		return "error", NewGQLError("删除料号尺寸数据失败，删除操作被终止", err.Error())
	}

	var pointIDs []int
	if err := tx.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Pluck("id", &pointIDs).Error; err != nil {
		return "error", NewGQLError("删除料号尺寸点位数据失败，删除操作被终止", err.Error())
	}

	if err := tx.Where("id = ?", id).Delete(orm.Material{}).Error; err != nil {
		return "error", NewGQLError("删除料号失败，发生了一些错误", err.Error())
	}

	if err := tx.Where("material_id = ?", id).Delete(orm.File{}).Error; err != nil {
		return "error", NewGQLError("删除料号失败，发生了一些错误", err.Error())
	}

	if err := tx.Where("material_id = ?", id).Delete(orm.Device{}).Error; err != nil {
		return "error", NewGQLError("删除料号设备失败，发生了一些错误", err.Error())
	}

	tx.Commit()

	go func() {
		orm.DB.Where("material_id = ?", id).Delete(orm.Product{})
		orm.DB.Where("point_id in (?)", pointIDs).Delete(orm.PointValue{})
		orm.DB.Where("id in (?)", pointIDs).Delete(orm.Point{})
		orm.DB.Where("id in (?)", sizeIDs).Delete(orm.Size{})
	}()

	return "料号删除成功", nil
}

func (r *queryResolver) MaterialYieldTop(ctx context.Context, duration []*time.Time, limit int) (*model.EchartsResult, error) {
	query := orm.DB.Model(&orm.Product{}).Select(
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
		return nil, NewGQLError("统计数据时发生了错误。", err.Error())
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
		return nil, NewGQLError("统计数据时发生了错误。", err.Error())
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

func (r *queryResolver) GroupAnalyzeMaterial(ctx context.Context, analyzeInput model.GroupAnalyzeInput) (*model.EchartsResult, error) {
	return groupAnalyze(analyzeInput, "material")
}
