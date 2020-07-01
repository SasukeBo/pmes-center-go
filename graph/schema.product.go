package graph

import (
	"context"
	"fmt"
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

func (r *queryResolver) Products(ctx context.Context, searchInput model.Search, page *int, limit int, offset *int) (*model.ProductWrap, error) {
	if searchInput.MaterialID == nil {
		return nil, NewGQLError("料号ID不能为空", "searchInput.MaterialID is nil")
	}
	oset := 0
	if offset != nil {
		oset = *offset
	} else if page != nil {
		if *page < 1 {
			return nil, NewGQLError("页数不能小于1", "")
		}
		oset = (*page - 1) * limit
	}

	var conditions []string
	var vars []interface{}
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return nil, NewGQLError("您所查找的料号不存在", fmt.Sprintf("get material with id = %v failed", *searchInput.MaterialID))
	}

	end := searchInput.EndTime
	if end == nil {
		t := time.Now()
		end = &t
	}
	begin := searchInput.BeginTime
	if begin == nil {
		t := end.AddDate(-1, 0, 0)
		begin = &t
	}

	// TODO: 关闭自动拉取
	//fileIDs, err := logic.FetchData(material, begin, end)
	//if err != nil {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(false), Message: stringP(err.Error())}
	//	return &model.ProductWrap{Status: status}, nil
	//}
	//if len(fileIDs) > 0 {
	//	status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
	//	return &model.ProductWrap{Status: status}, nil
	//}

	conditions = append(conditions, "material_id = ?")
	vars = append(vars, material.ID)
	if searchInput.DeviceID != nil {
		device := orm.GetDeviceWithID(*searchInput.DeviceID)
		if device != nil {
			conditions = append(conditions, "device_id = ?")
			vars = append(vars, device.ID)
		}
	}
	conditions = append(conditions, "created_at < ?")
	vars = append(vars, end)
	conditions = append(conditions, "created_at > ?")
	vars = append(vars, begin)

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

	fmt.Println(conditions)
	cond := strings.Join(conditions, " AND ")
	var products []orm.Product
	var out model.ProductWrap
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("id asc").Offset(oset).Limit(limit).Find(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound { // 无数据
			return &out, nil
		}

		return nil, NewGQLError("获取数据失败，请重试", err.Error())
	}

	var total int
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Count(&total).Error; err != nil {
		return nil, NewGQLError("统计产品数量失败", err.Error())
	}

	var productUUIDs []string
	for _, p := range products {
		productUUIDs = append(productUUIDs, p.UUID)
	}

	rows, err := orm.DB.Raw(`
	SELECT pv.product_uuid, p.name, pv.v FROM point_values AS pv
	JOIN points AS p ON pv.point_id = p.id
	WHERE pv.product_uuid IN (?)
	ORDER BY pv.product_uuid, p.index
	`, productUUIDs).Rows()
	if err != nil {
		return nil, NewGQLError("获取产品尺寸数据失败", err.Error())
	}
	defer rows.Close()

	var uuid, name string
	var value float64
	productPointValueMap := make(map[string]map[string]interface{})
	for rows.Next() {
		rows.Scan(&uuid, &name, &value)
		if p, ok := productPointValueMap[uuid]; ok {
			p[name] = value
			continue
		}

		productPointValueMap[uuid] = map[string]interface{}{name: value}
	}

	var outs []*model.Product
	for _, product := range products {
		var out model.Product
		if err := copier.Copy(&out, &product); err != nil {
			continue
		}

		if mp, ok := productPointValueMap[product.UUID]; ok {
			out.PointValue = mp
		}
		outs = append(outs, &out)
	}

	var sizeIDs []int
	orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs)

	var pointNames []string
	orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index asc").Pluck("name", &pointNames)

	return &model.ProductWrap{
		TableHeader: pointNames,
		Products:    outs,
		Total:       total,
	}, nil
}

func (r *queryResolver) ExportProducts(ctx context.Context, searchInput model.Search) (string, error) {
	if searchInput.MaterialID == nil {
		return "error", NewGQLError("料号ID不能为空")
	}
	material := orm.GetMaterialWithID(*searchInput.MaterialID)
	if material == nil {
		return "error", NewGQLError("料号不存在")
	}

	// 拼接查询条件
	var conditions []string
	var vars []interface{}
	conditions = append(conditions, "material_id = ?")
	vars = append(vars, material.ID)

	if searchInput.EndTime == nil {
		t := time.Now()
		searchInput.EndTime = &t
	}

	if searchInput.BeginTime == nil {
		t := searchInput.EndTime.AddDate(-1, 0, 0)
		searchInput.BeginTime = &t
	}
	conditions = append(conditions, "created_at < ?")
	vars = append(vars, searchInput.EndTime)
	conditions = append(conditions, "created_at > ?")
	vars = append(vars, searchInput.BeginTime)

	if searchInput.DeviceID != nil {
		device := orm.GetDeviceWithID(*searchInput.DeviceID)
		if device != nil {
			conditions = append(conditions, "device_id = ?")
			vars = append(vars, device.ID)
		}
	}

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

	opID := uuid.New().String()
	condition := strings.Join(conditions, " AND ")
	go logic.HandleExport(opID, material, searchInput, condition, vars...)

	return opID, nil
}

func (r *queryResolver) ExportFinishPercent(ctx context.Context, opID string) (*model.ExportResponse, error) {
	rsp, err := logic.CheckExport(opID)
	if err != nil {
		return nil, NewGQLError(rsp.Message, err.Error())
	}

	return rsp, nil
}

func (r *mutationResolver) CancelExport(ctx context.Context, opID string) (string, error) {
	err := logic.CancelExport(opID)
	if err != nil {
		return "error", NewGQLError("取消导出失败", err.Error())
	}

	return "ok", nil
}

func (r *queryResolver) ImportRecords(ctx context.Context, materialID, page, limit int) (*model.ImportRecordsWrap, error) {
	query := orm.DB.Model(&orm.File{}).Where("material_id = ?", materialID)
	var records []orm.File
	if err := query.Offset((page - 1) * limit).Limit(limit).Find(&records).Error; err != nil {
		return nil, NewGQLError("获取数据失败", err.Error())
	}
	var count int
	if err := query.Count(&count).Error; err != nil {
		return nil, NewGQLError("获取数据失败", err.Error())
	}

	var outs []*model.ImportRecord
	for _, record := range records {
		var out model.ImportRecord

		if err := copier.Copy(&out, &record); err != nil {
			continue
		}

		outs = append(outs, &out)
	}

	return &model.ImportRecordsWrap{
		Total:   count,
		Records: outs,
	}, nil
}

func (r *mutationResolver) RevertRecord(ctx context.Context, id int) (string, error) {
	record := orm.File{ID: id}
	orm.DB.Delete(&record)

	var uuids []string
	if err := orm.DB.Model(&orm.Product{}).Where("file_id = ?", id).Pluck("uuid", &uuids).Error; err != nil {
		return "error", err
	}

	orm.DB.Exec("DELETE FROM products WHERE uuid IN (?)", uuids)
	orm.DB.Exec("DELETE FROM point_values WHERE product_uuid IN (?)", uuids)

	return "ok", nil
}

func (r *mutationResolver) ImportData(ctx context.Context, materialID int, path string) (string, error) {
	var material orm.Material
	if err := orm.DB.Model(orm.Material{}).Where("id = ?", materialID).Find(&material).Error; err != nil {
		return "", NewGQLError("获取料号信息失败", err.Error())
	}

	if !ftpclient.CheckFile(path) {
		return "", NewGQLError("无法访问该文件，请确认路径是否正确", "")
	}

	if err := logic.FetchFileWithPath(material, path); err != nil {
		return "", NewGQLError("获取数据失败", err.Error())
	}

	return "ok", nil
}

func (r *queryResolver) CheckImport(ctx context.Context, id int) (*model.ImportResponse, error) {
	var file orm.File
	if err := orm.DB.Model(&orm.File{}).Where("id = ?", id).Find(&file).Error; err != nil {
		return nil, NewGQLError("获取导入记录失败", err.Error())
	}

	return &model.ImportResponse{
		Finished:     file.Finished,
		FinishedRows: file.FinishedRows,
	}, nil
}
