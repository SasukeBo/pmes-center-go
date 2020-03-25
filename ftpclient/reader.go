package ftpclient

// 编写读取ftp服务器数据的逻辑代码
// 解析FTP服务器文件
// 生成批量插入的SQL

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"path/filepath"
	"strconv"
	"strings"
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
	Index int
	USL float64
	LSL float64
}

type XLSXReader struct {
	DateSet [][]string           // only cache data of sheet 1
	DimSL   map[string]SL // map cache key (dim) and value([uSL, lSL])
	MaterialID string
	DeviceName string
	ProductUUIDPrefix string
}

func NewXLSXReader() *XLSXReader {
	return &XLSXReader{
		DateSet: make([][]string, 0),
		DimSL: make(map[string]SL),
	}
}

func (xr *XLSXReader) ReadSize(path string) error {
	dataSheet, err := read(path)
	if err != nil {
		return err
	}

	var dimSet, USLSet, LSLSet *[]string
	for i, row := range dataSheet {
		if row[0] == "Dim" {
			dimSet = &dataSheet[i]
			continue
		}
		if row[0] == "USL" {
			USLSet = &dataSheet[i]
			continue
		}
		if row[0] == "LSL" {
			LSLSet = &dataSheet[i]
			continue
		}
		if dimSet != nil && USLSet != nil && LSLSet != nil {
			break
		}
	}

	if dimSet == nil || USLSet == nil || LSLSet == nil {
		return errors.New("xlsx文件格式有误。")
	}

	for i, k := range *dimSet {
		if k == "" || k == "TEMP" || k == "Dim" {
			continue
		}
		xr.DimSL[k] = SL{
			USL: parseFloat((*USLSet)[i]),
			LSL: parseFloat((*LSLSet)[i]),
			Index: i,
		}
	}

	return nil
}

func (xr *XLSXReader) Read(path string) error {
	result := reg.FindAllStringSubmatch(filepath.Base(path), -1)
	if len(result) > 0 && len(result[0]) > 3 {
		xr.MaterialID = result[0][1]
		xr.DeviceName = fmt.Sprintf("%s设备%s", xr.MaterialID, result[0][2])
	} else {
		return &FTPError{
			Message: fmt.Sprintf("文件名格式不正确，%s", path),
		}
	}
	s1 := strings.Replace(filepath.Base(path), ".xlsx", "", -1)
	xr.ProductUUIDPrefix = strings.Replace(s1, "-", "", -1)
	dataSheet, err := read(path)
	if err != nil {
		return err
	}
	var dimSet, USLSet, LSLSet *[]string
	for i := 0; i < len(dataSheet); i++ {
		if dataSheet[i][0] == "Dim" {
			dimSet = &dataSheet[i]
			continue
		}

		if dataSheet[i][0] == "USL" {
			USLSet = &dataSheet[i]
			continue
		}

		if dataSheet[i][0] == "LSL" {
			LSLSet = &dataSheet[i]
			continue
		}

		if _, err := strconv.ParseInt(dataSheet[i][0], 10, 64); err == nil {
			xr.DateSet = dataSheet[i:]
			break
		}
	}

	if dimSet == nil || USLSet == nil || LSLSet == nil {
		return errors.New("xlsx文件格式有误。")
	}

	for i, k := range *dimSet {
		if k == "" || k == "TEMP" || k == "Dim" {
			continue
		}
		xr.DimSL[k] = SL{
			USL: parseFloat((*USLSet)[i]),
			LSL: parseFloat((*LSLSet)[i]),
			Index: i,
		}
	}

	return nil
}

func read(path string) ([][]string, error) {
	content, err := ReadFile(path)
	if err != nil {
		if fe, ok := err.(*FTPError); ok {
			fe.Logger()
			return nil, err
		}

		log.Println(err)
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

