package logic

import (
	"context"
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/tealeg/xlsx/v3"
	"path/filepath"
	"strconv"
)

func LoadFile(ctx context.Context, id uint) *model.File {
	var file orm.File
	if err := file.Get(id); err != nil {
		return nil
	}

	var out model.File
	if err := copier.Copy(&out, &file); err != nil {
		return nil
	}

	return &out
}

func assembleDataIntoFile(name string, points []orm.Point, data []orm.Product) (*orm.File, error) {
	xlsxObj := xlsx.NewFile()
	sheet, err := xlsxObj.AddSheet("data")
	if err != nil {
		return nil, err
	}
	var labelCellLength = points[0].Index - 2
	var row *xlsx.Row
	{
		// 第一行
		row = addRow(sheet)
		{
			labelCell := row.AddCell()
			labelCell.SetString("Dim")
			for i := 0; i < labelCellLength; i++ {
				cell := row.AddCell()
				addLabelCellStyle(cell)
			}
			labelCell.Merge(labelCellLength, 0)
			addLabelCellStyle(labelCell)

			for _, p := range points {
				pointCell := row.AddCell()
				pointCell.SetString(p.Name)
				addPointCellStyle(pointCell)
			}
		}
		// 第二行
		row = addRow(sheet)
		{
			labelCell := row.AddCell()
			labelCell.SetString("USL")
			for i := 0; i < labelCellLength; i++ {
				cell := row.AddCell()
				addLabelCellStyle(cell)
			}
			labelCell.Merge(labelCellLength, 0)
			addLabelCellStyle(labelCell)

			for _, p := range points {
				pointCell := row.AddCell()
				pointCell.SetFloat(p.UpperLimit)
				addPointCellStyle(pointCell)
			}
		}
		// 第三行
		row = addRow(sheet)
		{
			labelCell := row.AddCell()
			labelCell.SetString("NOM")
			for i := 0; i < labelCellLength; i++ {
				cell := row.AddCell()
				addLabelCellStyle(cell)
			}
			labelCell.Merge(labelCellLength, 0)
			addLabelCellStyle(labelCell)

			for _, p := range points {
				pointCell := row.AddCell()
				pointCell.SetFloat(p.Nominal)
				addPointCellStyle(pointCell)
			}
		}
		// 第四行
		row = addRow(sheet)
		{
			labelCell := row.AddCell()
			labelCell.SetString("LSL")
			for i := 0; i < labelCellLength; i++ {
				cell := row.AddCell()
				addLabelCellStyle(cell)
			}
			labelCell.Merge(labelCellLength, 0)
			addLabelCellStyle(labelCell)

			for _, p := range points {
				pointCell := row.AddCell()
				pointCell.SetFloat(p.LowerLimit)
				addPointCellStyle(pointCell)
			}
		}
		// 第五行
		row = addRow(sheet)
		{
			cell := row.AddCell()
			cell.SetString("序号")
			addLabelCellStyle(cell)
			cell = row.AddCell()
			cell.SetString("检测时间")
			addLabelCellStyle(cell)
			cell = row.AddCell()
			cell.SetString("条码号")
			addLabelCellStyle(cell)
			for i := 0; i < labelCellLength-2; i++ {
				cell = row.AddCell()
				addLabelCellStyle(cell)
			}
		}
	}

	// 填入数据
	{
		for i, p := range data {
			row = addRow(sheet)
			cell := row.AddCell()
			cell.SetInt(i)
			addBasicCellStyle(cell)
			cell = row.AddCell()
			cell.SetDateTime(p.CreatedAt)
			addBasicCellStyle(cell)
			cell = row.AddCell()
			cell.SetString(p.BarCode)
			addBasicCellStyle(cell)
			for i := 0; i < labelCellLength-2; i++ {
				cell = row.AddCell()
				addBasicCellStyle(cell)
			}
			for _, point := range points {
				cell = row.AddCell()
				var value float64
				if v, ok := p.PointValues[point.Name]; !ok {
					value = 0
				} else {
					value, _ = strconv.ParseFloat(fmt.Sprint(v), 64)
				}
				cell.SetFloat(value)
				addValueCellStyle(cell)
			}
		}
	}
	// 设置列属性
	col := xlsx.NewColForRange(0, 0)
	col.SetWidth(10)
	sheet.SetColParameters(col)

	col = xlsx.NewColForRange(1, 1)
	col.SetWidth(25)
	sheet.SetColParameters(col)

	col = xlsx.NewColForRange(2, 2)
	col.SetWidth(35)
	sheet.SetColParameters(col)

	col = xlsx.NewColForRange(3, points[len(points)-1].Index)
	col.SetWidth(10)
	sheet.SetColParameters(col)

	// 保存文件
	dst := configer.GetString("file_cache_path")
	token, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	//var relevantPath = filepath.Join(orm.DirCache, token.String())
	var relevantPath = filepath.Join(orm.DirCache, "test.xlsx")
	path := filepath.Join(dst, relevantPath)
	if err = xlsxObj.Save(path); err != nil {
		return nil, err
	}

	var file = &orm.File{
		Name:        name,
		Path:        relevantPath,
		Token:       token.String(),
		ContentType: orm.XlsxContentType,
	}
	if err = orm.Save(file).Error; err != nil {
		return nil, err
	}

	fmt.Println(path)
	return file, nil
}

func addCellStyle(style *xlsx.Style) {
	style.ApplyBorder = true
	style.ApplyFill = true
	style.ApplyFont = true
	style.ApplyAlignment = true
	style.Border = xlsx.Border{
		Left:   "thin",
		Right:  "thin",
		Top:    "thin",
		Bottom: "thin",
	}
	style.Alignment = xlsx.Alignment{
		Horizontal: "center",
		Vertical:   "center",
	}
	style.Font = xlsx.Font{
		Size: 10,
		Name: "Verdana",
	}
}

func addLabelCellStyle(cell *xlsx.Cell) {
	style := cell.GetStyle()
	if style == nil {
		style = xlsx.NewStyle()
		cell.SetStyle(style)
	}
	addCellStyle(style)
	style.Fill = xlsx.Fill{
		PatternType: "solid",
		BgColor:     "FF333333",
		FgColor:     "FFFEFF00",
	}
}

func addPointCellStyle(cell *xlsx.Cell) {
	style := cell.GetStyle()
	if style == nil {
		style = xlsx.NewStyle()
		cell.SetStyle(style)
	}
	addCellStyle(style)
	style.Fill = xlsx.Fill{
		PatternType: "solid",
		BgColor:     "FF333333",
		FgColor:     "FFD9D9D9",
	}
}

func addBasicCellStyle(cell *xlsx.Cell) {
	style := cell.GetStyle()
	if style == nil {
		style = xlsx.NewStyle()
		cell.SetStyle(style)
	}
	addCellStyle(style)
}

func addValueCellStyle(cell *xlsx.Cell) {
	style := cell.GetStyle()
	if style == nil {
		style = xlsx.NewStyle()
		cell.SetStyle(style)
	}
	addCellStyle(style)
	style.Fill = xlsx.Fill{
		PatternType: "solid",
		BgColor:     "FF333333",
		FgColor:     "FF00AA32",
	}
}

func addRow(sheet *xlsx.Sheet) *xlsx.Row {
	row := sheet.AddRow()
	row.SetHeight(13.2)
	return row
}
