package logic

import (
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/tealeg/xlsx"
)

const (
	CELL_RGB_COLOR_LIGHT_GREEN = "00EFFFEF"
	CELL_RGB_COLOR_YELLOW      = "7DFFFF00"
	CELL_RGB_COLOR_DARK_GREEN  = "00008000"
	CELL_RGB_COLOR_MEAT_YELLOW = "00FFFFE7"
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
	subHeaderCellStyle      = newNormalStyle(CELL_RGB_COLOR_YELLOW)
	headerCellStyle = newNormalStyle(CELL_RGB_COLOR_LIGHT_GREEN)
)

type rowMap map[string]*xlsx.Row

func CreateXLSXHeader(sheet *xlsx.Sheet, points []orm.Point) {
	// table header
	rMap := make(rowMap)
	rowLabels := []string{"Dim", "USL", "Nominal", "LSL"}
	for _, label := range rowLabels {
		rMap[label] = genRowWithLabelCell(sheet, label)
	}

	for _, p := range points {
		dimRow := rMap["Dim"]
		dimCell := dimRow.AddCell()
		dimCell.SetStyle(headerCellStyle)
		dimCell.SetString(p.Name)

		uslRow := rMap["USL"]
		uslCell := uslRow.AddCell()
		uslCell.SetStyle(headerCellStyle)
		uslCell.SetValue(p.UpperLimit)

		nominalRow := rMap["Nominal"]
		nominalCell := nominalRow.AddCell()
		nominalCell.SetStyle(headerCellStyle)
		nominalCell.SetValue(p.Norminal)

		lslRow := rMap["LSL"]
		lslCell := lslRow.AddCell()
		lslCell.SetStyle(headerCellStyle)
		lslCell.SetValue(p.LowerLimit)
	}
}

func CreateXLSXSumRows(sheet *xlsx.Sheet) rowMap {
	rMap := make(rowMap)
	rowNames := []string{"Total Qty", "OK Qty", "NG Qty", "Yield", "Mean", "Cp", "Cpk"}
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
