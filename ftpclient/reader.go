package ftpclient

// 编写读取ftp服务器数据的逻辑代码
// 解析FTP服务器文件
// 生成批量插入的SQL

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/tealeg/xlsx"
)

type XLSXReader struct {
	DataSet        [][]string          // only cache data of sheet 1
	Material       *orm.Material       // 导入的料号
	Device         *orm.Device         // 导入设备
	Record         *orm.ImportRecord   // 导入记录
	DecodeTemplate *orm.DecodeTemplate // 解析模板
}

func NewXLSXReader(material *orm.Material, device *orm.Device, template *orm.DecodeTemplate) *XLSXReader {
	return &XLSXReader{
		DataSet:        make([][]string, 0),
		Material:       material,
		Device:         device,
		DecodeTemplate: template,
	}
}

func (xr *XLSXReader) Read(path string) error {
	dataSheet, err := read(path)
	if err != nil {
		return err
	}
	var bIdx = xr.DecodeTemplate.DataRowIndex
	var eIdx = len(dataSheet) - 1
	for i, row := range dataSheet {
		if (len(row) == 0 || row[0] == "") && i > bIdx { // 空行 或 行首位空 为截至行，理想情况下不存在数据行中穿插空行
			eIdx = i - 1
			break
		}
	}
	dataSet := dataSheet[bIdx : eIdx+1]
	xr.DataSet = dataSet

	log.Info("data begin idx: %v, end idx: %v\n", bIdx, eIdx)
	return nil
}

func read(path string) ([][]string, error) {
	content, err := ReadFile(path)
	if err != nil {
		if fe, ok := err.(*FTPError); ok {
			fe.Logger()
			return nil, err
		}

		log.Errorln(err)
		return nil, &FTPError{
			Message:   fmt.Sprintf("从FTP服务器读取文件%s失败", path),
			OriginErr: err,
		}
	}

	file, err := xlsx.OpenBinary(content)
	if err != nil {
		return nil, fmt.Errorf("读取数据文件失败，原始错误信息: %v", err)
	}

	originData, err := file.ToSlice()
	if err != nil {
		return nil, err
	}
	if len(originData) == 0 {
		return nil, errors.New("xlsx文件内容是空的。")
	}

	return originData[0], nil
}
