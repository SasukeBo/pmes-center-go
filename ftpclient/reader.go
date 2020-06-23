package ftpclient

// 编写读取ftp服务器数据的逻辑代码
// 解析FTP服务器文件
// 生成批量插入的SQL

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/SasukeBo/log"
	"github.com/tealeg/xlsx"
	"path/filepath"
	"strconv"
)

// CSVDecoder csv decoder object
type CSVDecoder struct {
	Headers    []string
	Rows       [][]string
	Limits     [][]string
	MaterialID string
	DeviceName string
}

// Decode 解析csv文件，输出二维字符串
func (cd *CSVDecoder) Decode(data []byte) error {
	reader := csv.NewReader(bytes.NewReader(data))
	rr, err := reader.ReadAll()
	if err != nil {
		return err
	}

	cd.Headers = rr[0]
	cd.Limits = rr[1:3]
	cd.Rows = rr[3:]
	return nil
}

type SL struct {
	Index   int
	USL     float64
	Nominal float64
	LSL     float64
}

type XLSXReader struct {
	DataSet    [][]string    // only cache data of sheet 1
	DimSL      map[string]SL // map cache key (dim) and value([uSL, lSL])
	MaterialID string
	DeviceName string
	PathID     int // 读取文件路径id
}

func NewXLSXReader() *XLSXReader {
	return &XLSXReader{
		DataSet: make([][]string, 0),
		DimSL:   make(map[string]SL),
	}
}

func (xr *XLSXReader) ReadSize(path string) error {
	dataSheet, err := read(path)
	if err != nil {
		return err
	}

	var dimSet, USLSet, LSLSet *[]string
	for i, row := range dataSheet {
		if len(row) == 0 {
			continue
		}
		switch row[0] {
		case "Dim":
			dimSet = &dataSheet[i]
		case "USL":
			USLSet = &dataSheet[i]
		case "LSL":
			LSLSet = &dataSheet[i]
		}

		if dimSet != nil && USLSet != nil && LSLSet != nil {
			break
		}
	}

	if dimSet == nil || USLSet == nil || LSLSet == nil {
		err := errors.New(fmt.Sprintf("xlsx文件格式有误。%s", path))
		log.Errorln(err)
		return err
	}

	for i, k := range *dimSet {
		if k == "" || k == "TEMP" || k == "Dim" {
			continue
		}
		usl := parseFloat((*USLSet)[i])
		lsl := parseFloat((*LSLSet)[i])
		xr.DimSL[k] = SL{
			USL:     usl,
			Nominal: (usl + lsl) / 2,
			LSL:     lsl,
			Index:   i,
		}
	}

	return nil
}

func (xr *XLSXReader) Read(path string) error {
	result := reg.FindAllStringSubmatch(filepath.Base(path), -1)
	if len(result) > 0 && len(result[0]) > 3 {
		xr.MaterialID = result[0][1]
		xr.DeviceName = result[0][2]
	} else {
		return &FTPError{
			Message: fmt.Sprintf("文件名格式不正确，%s", path),
		}
	}
	dataSheet, err := read(path)
	if err != nil {
		return err
	}
	var bIdx = 0
	var eIdx = 0
	for i, row := range dataSheet {
		if len(row) == 0 {
			eIdx = i
			continue
		}
		if bIdx == 0 {
			_, err := strconv.Atoi(row[0])
			if err == nil {
				bIdx = i
			}
		}
		if bIdx > 0 && eIdx > 0 && eIdx > bIdx {
			break
		}
	}
	if eIdx == 0 {
		eIdx = len(dataSheet)
	}
	dataSet := dataSheet[bIdx : eIdx-1]
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
