package logic

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/conf"
	"github.com/SasukeBo/ftpviewer/graph/model"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	CellRgbColorLightGreen = "00EFFFEF"
	CellRgbColorYellow     = "7DFFFF00"
	CellRgbColorDarkGreen  = "F9008000"
	CellRgbColorWhite      = "00FFFFFF"
	CellRgbColorRed        = "F9F50808"
)

const (
	CellNameTotalQty = "Total Qty"
	CellNameOKQty    = "OK Qty"
	CellNameNGQty    = "NG Qty"
	CellNameYield    = "Yield"
	CellNameMean     = "Mean"
	CellNameCP       = "Cp"
	CellNameCPK      = "Cpk"
	CellNameDim      = "Dim"
	CellNameUSL      = "USL"
	CellNameLSL      = "LSL"
	CellNameNominal  = "Nominal"
)

func newNormalStyle(rgb string) *xlsx.Style {
	style := xlsx.Style{
		Fill: xlsx.Fill{
			PatternType: "solid",
			FgColor:     rgb,
		},
		Border: xlsx.Border{
			Left:   "thin",
			Right:  "thin",
			Top:    "thin",
			Bottom: "thin",
		},
		Alignment: xlsx.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	}
	return &style
}

var (
	subHeaderCellStyle = newNormalStyle(CellRgbColorYellow)
	headerCellStyle    = newNormalStyle(CellRgbColorLightGreen)
	errorRowCellStyle  = newNormalStyle(CellRgbColorRed)
	normalCellStyle    = newNormalStyle(CellRgbColorWhite)
	dataCellStyle      = newNormalStyle(CellRgbColorDarkGreen)
)

type rowMap map[string]*xlsx.Row

func CreateXLSXHeader(sheet *xlsx.Sheet, points []orm.Point) {
	// table header
	rMap := make(rowMap)
	rowLabels := []string{CellNameDim, CellNameUSL, CellNameNominal, CellNameLSL}
	for _, label := range rowLabels {
		rMap[label] = genRowWithLabelCell(sheet, label)
	}

	for _, p := range points {
		dimRow := rMap[CellNameDim]
		dimCell := dimRow.AddCell()
		dimCell.SetStyle(headerCellStyle)
		dimCell.SetString(p.Name)

		uslRow := rMap[CellNameUSL]
		uslCell := uslRow.AddCell()
		uslCell.SetStyle(headerCellStyle)
		uslCell.SetValue(p.UpperLimit)

		nominalRow := rMap[CellNameNominal]
		nominalCell := nominalRow.AddCell()
		nominalCell.SetStyle(headerCellStyle)
		nominalCell.SetValue(p.Norminal)

		lslRow := rMap[CellNameLSL]
		lslCell := lslRow.AddCell()
		lslCell.SetStyle(headerCellStyle)
		lslCell.SetValue(p.LowerLimit)
	}
}

func CreateXLSXSumRows(sheet *xlsx.Sheet) rowMap {
	rMap := make(rowMap)
	rowNames := []string{CellNameTotalQty, CellNameOKQty, CellNameNGQty, CellNameYield, CellNameMean, CellNameCP, CellNameCPK}
	for _, name := range rowNames {
		rMap[name] = genRowWithLabelCell(sheet, name)
	}

	return rMap
}

func genRowWithLabelCell(sheet *xlsx.Sheet, label string) *xlsx.Row {
	row := sheet.AddRow()
	labelCell := row.AddCell()
	labelCell.Merge(6, 0)
	row.AddCell().SetStyle(headerCellStyle)
	row.AddCell().SetStyle(headerCellStyle)
	row.AddCell().SetStyle(headerCellStyle)
	row.AddCell().SetStyle(headerCellStyle)
	row.AddCell().SetStyle(headerCellStyle)
	row.AddCell().SetStyle(headerCellStyle)
	labelCell.SetString(label)
	labelCell.SetStyle(headerCellStyle)
	return row
}

func CreateXLSXSubHeader(sheet *xlsx.Sheet) {
	subHeaderRow := sheet.AddRow()
	cellValues := []string{"NO.", "日期", "2D条码号", "线体号", "冶具号", "模号", "班别"}
	for _, v := range cellValues {
		cell := subHeaderRow.AddCell()
		cell.SetStyle(subHeaderCellStyle)
		cell.SetValue(v)
	}
}

type handlerResponse struct {
	err        error   // 处理错误
	percent    float64 // 阶段完成百分比
	message    string  // 阶段描述
	fileName   string  // 生成的文件名称
	finished   bool    // 是否已完成
	cancelChan chan struct{}
}

var handlerCache map[string]*handlerResponse

const (
	xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	pvSQL           = `SELECT pv.v FROM point_values AS pv JOIN points AS p ON pv.point_id = p.id WHERE pv.product_uuid = ? ORDER BY p.index ASC`
)

func HandleExport(opID string, material *orm.Material, search model.Search, condition string, vars ...interface{}) {
	response := &handlerResponse{
		message:    "正在准备导出数据",
		cancelChan: make(chan struct{}, 0),
	}
	handlerCache[opID] = response

	// 创建文件
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("data")
	if err != nil {
		response.err = err
		response.message = "导出失败，发生了一些错误"
		return
	}

	// 获取表头信息
	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
		response.err = err
		response.message = "查询数据时发生错误，导出失败"
		return
	}

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Order("points.index ASC").Find(&points).Error; err != nil {
		response.err = err
		response.message = "查询数据时发生错误，导出失败"
		return
	}

	// 写入头部数据
	CreateXLSXHeader(sheet, points)
	rMap := CreateXLSXSumRows(sheet)
	CreateXLSXSubHeader(sheet)

	// 开始处理数据
	var total, finished float64
	var products []orm.Product
	if err := orm.DB.Model(&orm.Product{}).Where(condition, vars...).Order("id ASC").Find(&products).Error; err != nil {
		response.err = err
		response.message = "查询数据时发生错误，导出失败"
		return
	}

	response.message = "正在处理数据"
	total = float64(len(products))
	pChan := make(chan orm.Product, 1)
	finishChan := make(chan struct{}, 0)
	go func() {
		for _, p := range products {
			pChan <- p
		}
		close(finishChan)
	}()

	stopLoop := false
	for {
		select {
		case <-finishChan:
			stopLoop = true
			break

		case <-response.cancelChan:
			response.message = "已取消"
			response.finished = true
			return

		case p := <-pChan:
			row := sheet.AddRow()
			hds := normalCellStyle
			bds := dataCellStyle
			if !p.Qualified {
				hds = errorRowCellStyle
				bds = errorRowCellStyle
			}
			appendValueWithFgColor(row, hds, p.ID)
			appendValueWithFgColor(row, hds, p.CreatedAt.Format("2006-01-02T15:04:05"))
			appendValueWithFgColor(row, hds, p.D2Code)
			appendValueWithFgColor(row, hds, p.LineID)
			appendValueWithFgColor(row, hds, p.JigID)
			appendValueWithFgColor(row, hds, p.MouldID)
			appendValueWithFgColor(row, hds, p.ShiftNumber)

			sqlRows, err := orm.DB.Raw(pvSQL, p.UUID).Rows()
			if err != nil {
				continue
			}
			for sqlRows.Next() {
				var pv float64
				sqlRows.Scan(&pv)
				appendValueWithFgColor(row, bds, pv)
			}

			sqlRows.Close()
			finished++
			response.percent = finished / total
		}

		if stopLoop {
			break
		}
	}

	response.message = "正在处理统计数据"
	xfSlice, err := file.ToSlice()
	if err != nil {
		response.err = err
		response.message = "统计数据时发生错误，导出失败"
	}
	dataRows := xfSlice[0]
	for i, p := range points {
		pvs := make([]float64, 0)
		for j := 12; j < len(dataRows); j++ {
			v := dataRows[j][i+7]
			pv, _ := strconv.ParseFloat(v, 64)
			pvs = append(pvs, pv)
		}
		calculateAndCreate(rMap, p, pvs)
	}

	response.message = "正在写入文件"
	fileNameParts := []string{material.Name}

	if search.DeviceID != nil {
		device := orm.GetDeviceWithID(*search.DeviceID)
		if device != nil {
			fileNameParts = append(fileNameParts, device.Name)
		}
	}

	fileNameParts = append(fileNameParts, search.BeginTime.Format("20060102"))
	fileNameParts = append(fileNameParts, search.EndTime.Format("20060102"))
	fileName := strings.Join(fileNameParts, "-") + ".xlsx"
	filePath := filepath.Join(conf.FileCachePath, fileName)

	// 输出文件
	file.Save(filePath)
	response.fileName = fileName
	response.finished = true
	response.message = "导出成功"
}

func calculateAndCreate(rMap rowMap, point orm.Point, values []float64) {
	_, cp, cpk, avg, ok, total, _ := AnalyzePointValues(point, values)
	appendValueWithFgColor(rMap[CellNameTotalQty], headerCellStyle, total)
	appendValueWithFgColor(rMap[CellNameOKQty], headerCellStyle, ok)
	appendValueWithFgColor(rMap[CellNameNGQty], headerCellStyle, total-ok)

	yield := math.Round(float64(ok)/float64(total)*10000) / 100
	appendValueWithFgColor(rMap[CellNameYield], headerCellStyle, fmt.Sprintf("%v%%", yield))
	avg = math.Round(avg*1000) / 1000
	appendValueWithFgColor(rMap[CellNameMean], headerCellStyle, avg)
	cp = math.Round(cp*100) / 100
	appendValueWithFgColor(rMap[CellNameCP], headerCellStyle, cp)
	cpk = math.Round(cpk*100) / 100
	appendValueWithFgColor(rMap[CellNameCPK], headerCellStyle, cpk)
}

func appendValueWithFgColor(row *xlsx.Row, style *xlsx.Style, v interface{}) {
	cell := row.AddCell()
	cell.SetStyle(style)
	cell.SetValue(v)
}

type object map[string]string

func Download(c *gin.Context) {
	fileName, ok := c.GetQuery("file_name")
	if !ok {
		c.JSON(http.StatusBadRequest, object{
			"message": "请求参数缺少文件名",
		})
		return
	}

	filePath := filepath.Join(conf.FileCachePath, fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, object{
			"message": "下载文件失败",
			"err":     err.Error(),
		})
	}
	os.Remove(filePath) // 删除临时文件
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, xlsxContentType, data)
}

func CheckExport(opID string) (*model.ExportResponse, error) {
	rsp, ok := handlerCache[opID]
	if !ok {
		return nil, errors.New("没有该导出任务的进度记录")
	}

	var fileName = rsp.fileName
	out := &model.ExportResponse{
		Percent:  rsp.percent,
		Message:  rsp.message,
		FileName: &fileName,
		Finished: rsp.finished,
	}
	if rsp.finished {
		delete(handlerCache, opID)
	}

	return out, rsp.err
}

func CancelExport(opID string) error {
	rsp, ok := handlerCache[opID]
	if !ok {
		return errors.New("没有该导出任务的进度记录")
	}

	close(rsp.cancelChan)
	delete(handlerCache, opID)
	return nil
}

func init() {
	handlerCache = make(map[string]*handlerResponse)
}
