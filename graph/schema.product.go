package graph

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"strconv"
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

	fileIDs, err := logic.NeedFetch(material, begin, end)
	if err != nil {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(false), Message: stringP(err.Error())}
		return &model.ProductWrap{Status: status}, nil
	}

	if len(fileIDs) > 0 {
		status := &model.FetchStatus{FileIDs: fileIDs, Pending: boolP(true), Message: stringP("需要从FTP服务器获取该时间段内料号数据")}
		return &model.ProductWrap{Status: status}, nil
	}

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
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("id asc").Offset(oset).Limit(limit).Find(&products).Error; err != nil {
		if err == gorm.ErrRecordNotFound { // 无数据
			return &model.ProductWrap{
				TableHeader: nil,
				Products:    nil,
				Status:      nil,
				Total:       intP(0),
			}, nil
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

	var outProducts []*model.Product
	for _, i := range products {
		p := i
		op := &model.Product{
			ID:          &p.ID,
			UUID:        &p.UUID,
			MaterialID:  &p.MaterialID,
			DeviceID:    &p.DeviceID,
			Qualified:   &p.Qualified,
			CreatedAt:   &p.CreatedAt,
			D2code:      &p.D2Code,
			LineID:      &p.LineID,
			JigID:       &p.JigID,
			MouldID:     &p.MouldID,
			ShiftNumber: &p.ShiftNumber,
		}
		if mp, ok := productPointValueMap[p.UUID]; ok {
			op.PointValue = mp
		}
		outProducts = append(outProducts, op)
	}

	var sizeIDs []int
	orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs)

	var pointNames []string
	orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index asc").Pluck("name", &pointNames)

	status := model.FetchStatus{Pending: boolP(false)}
	return &model.ProductWrap{
		TableHeader: pointNames,
		Products:    outProducts,
		Status:      &status,
		Total:       &total,
	}, nil
}

type object map[string]interface{}

// ExportsProducts 处理数据导出
func ExportsProducts(ctx *gin.Context) {
	materialID := ctx.Param("material_id")
	id, err := strconv.Atoi(materialID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, object{
			"message": "料号ID不是数字",
			"error":   err.Error(),
		})
		return
	}
	material := orm.GetMaterialWithID(id)
	if material == nil {
		ctx.JSON(http.StatusBadRequest, object{"message": "未找到料号"})
		return
	}
	// 创建文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("data")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, object{
			"message":   "导出数据失败，发生了一些错误",
			"originErr": err.Error(),
		})
		return
	}

	// 获取表头信息
	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, object{"message": "系统错误", "err": err.Error()})
		return
	}

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index asc").Find(&points).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, object{"message": "系统错误", "err": err.Error()})
		return
	}

	// 写入头部数据
	logic.CreateXLSXHeader(sheet, points)
	//rowMap := logic.CreateXLSXSumRows(sheet)
	logic.CreateXLSXSumRows(sheet)
	logic.CreateXLSXSubHeader(sheet)

	// 查询数据
	var conditions []string
	var vars []interface{}

	var end, begin *time.Time
	if endStr, ok := ctx.GetQuery("end_time"); ok {
		if t, err := time.Parse(time.RFC3339, endStr); err != nil {
			end = &t
		}
	}
	if end == nil {
		t := time.Now()
		end = &t
	}

	if beginStr, ok := ctx.GetQuery("begin_time"); ok {
		if t, err := time.Parse(time.RFC3339, beginStr); err != nil {
			begin = &t
		}
	}
	if begin == nil {
		t := end.AddDate(-1, 0, 0)
		begin = &t
	}

	conditions = append(conditions, "material_id = ?")
	vars = append(vars, material.ID)
	if v, ok := ctx.GetQuery("device_id"); ok {
		deviceID, err := strconv.Atoi(v)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, object{"message": "设备ID不是数字", "err": err.Error()})
			return
		}
		device := orm.GetDeviceWithID(deviceID)
		if device != nil {
			conditions = append(conditions, "device_id = ?")
			vars = append(vars, device.ID)
		}
	}

	conditions = append(conditions, "created_at < ?")
	vars = append(vars, end)
	conditions = append(conditions, "created_at > ?")
	vars = append(vars, begin)

	if lineID, ok := ctx.GetQuery("line_id"); ok {
		conditions = append(conditions, "line_id = ?")
		vars = append(vars, lineID)
	}

	if mouldID, ok := ctx.GetQuery("mould_id"); ok {
		conditions = append(conditions, "mould_id = ?")
		vars = append(vars, mouldID)
	}

	if jigID, ok := ctx.GetQuery("jig_id"); ok {
		conditions = append(conditions, "jig_id = ?")
		vars = append(vars, jigID)
	}

	if shiftNumber, ok := ctx.GetQuery("shift_number"); ok {
		conditions = append(conditions, "shift_number = ?")
		vars = append(vars, shiftNumber)
	}

	cond := strings.Join(conditions, " AND ")
	var products []orm.Product
	if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("id asc").Find(&products).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, object{"message": "导出数据失败，发生了一些错误", "err": err.Error()})
		return
	}


	var sql = `
		SELECT
			p.name,
			pv.v
		FROM
			point_values AS pv
			JOIN points AS p ON pv.point_id = p.id
		WHERE
			pv.product_uuid = ?
		ORDER BY
			pv.product_uuid, p.index
		`



	// 输出文件
	buf := bytes.NewBufferString("")
	file.Write(buf)

	xlsxContentType := "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ctx.DataFromReader(http.StatusOK, int64(buf.Len()), xlsxContentType, buf, map[string]string{"Content-Disposition": "attachment; filename=color-map.xlsx"})
}

func (r *queryResolver) ExportProducts(ctx context.Context, searchInput model.Search) (*model.Download, error) {
	/*
		if searchInput.MaterialID == nil {
			return nil, NewGQLError("料号ID不能为空", "searchInput.MaterialID is nil")
		}

		material := orm.GetMaterialWithID(*searchInput.MaterialID)
		if material == nil {
			return nil, NewGQLError("您所查找的料号不存在", fmt.Sprintf("get material with id = %v failed", *searchInput.MaterialID))
		}

		var conditions []string
		var vars []interface{}

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

		cond := strings.Join(conditions, " AND ")
		var products []orm.Product
		if err := orm.DB.Model(&orm.Product{}).Where(cond, vars...).Order("id asc").Find(&products).Error; err != nil {
			return nil, NewGQLError("导出数据失败，发生了一些错误", err.Error())
		}

		var sizeIDs []int
		orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs)

		var pointNames []string
		orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index asc").Pluck("name", &pointNames)

		var sql = `
		SELECT
			p.name,
			pv.v
		FROM
			point_values AS pv
			JOIN points AS p ON pv.point_id = p.id
		WHERE
			pv.product_uuid = ?
		ORDER BY
			pv.product_uuid, p.index
		`

		productPointValueMap := make(map[string]interface{})
		for _, p := range products {
			rows, err := orm.DB.Raw(sql, p.UUID).Rows()
			if err != nil {
				rows.Close()
				continue
			}

			var name string
			var value float64
			catch := make(map[string]interface{})
			for rows.Next() {
				rows.Scan(&name, &value)
				catch[name] = value
			}
			rows.Close()
			productPointValueMap[p.UUID] = catch
		}
	*/

	//return productPointValueMap, nil

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("data")
	if err != nil {
		return nil, NewGQLError("导出数据失败，发生了一些错误", err.Error())
	}
	row := sheet.AddRow()
	cell := row.AddCell()
	cell.SetString("hello world")

	buf := bytes.NewBufferString("")
	file.Write(buf)
	if err := file.Save("/Users/sasukebo/Downloads/test.xlsx"); err != nil {
		log.Errorln(err)
	}

	if err := ioutil.WriteFile("/Users/sasukebo/Downloads/test2.xlsx", buf.Bytes(), 0644); err != nil {
		log.Errorln(err)
	}

	content := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &model.Download{
		FileContent:   content,
		FileExtension: "xlsx",
	}, nil
}
