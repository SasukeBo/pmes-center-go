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
	"time"
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
	Index    int
	USL      float64
	Norminal float64
	LSL      float64
}

type XLSXReader struct {
	DateSet           [][]string    // only cache data of sheet 1
	DimSL             map[string]SL // map cache key (dim) and value([uSL, lSL])
	MaterialID        string
	DeviceName        string
	ProductUUIDPrefix string
	ProductAt         *time.Time
	PathID            int // 读取文件路径id
}

func NewXLSXReader() *XLSXReader {
	return &XLSXReader{
		DateSet: make([][]string, 0),
		DimSL:   make(map[string]SL),
	}
}

func (xr *XLSXReader) ReadSize(path string) error {
	dataSheet, err := read(path)
	if err != nil {
		return err
	}

	var dimSet, USLSet, LSLSet *[]string
	dimSet = &dataSheet[2]
	USLSet = &dataSheet[3]
	LSLSet = &dataSheet[5]

	if dimSet == nil || USLSet == nil || LSLSet == nil {
		return errors.New("xlsx文件格式有误。")
	}

	for i, k := range *dimSet {
		if k == "" || k == "TEMP" || k == "Dim" {
			continue
		}
		usl := parseFloat((*USLSet)[i])
		lsl := parseFloat((*LSLSet)[i])
		xr.DimSL[k] = SL{
			USL:      usl,
			Norminal: (usl + lsl) / 2,
			LSL:      lsl,
			Index:    i,
		}
	}

	return nil
}

func (xr *XLSXReader) Read(path string) error {
	result := reg.FindAllStringSubmatch(filepath.Base(path), -1)
	if len(result) > 0 && len(result[0]) > 3 {
		dateStr := result[0][3]
		year, _ := strconv.Atoi(dateStr[:4])
		month, _ := strconv.Atoi(dateStr[4:6])
		day, _ := strconv.Atoi(dateStr[6:])
		productAt := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

		xr.ProductAt = &productAt
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
	xr.DateSet = dataSheet[15:]
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

	fmt.Printf("total size of content: %v\n", len(content))
	file, err := xlsx.OpenBinary(content)
	fmt.Printf("total sheets count: %v\n", len(file.Sheets))
	for _, v := range file.Sheets {
		fmt.Printf("sheet %s has %v rows\n", v.Name, len(v.Rows))
	}
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
