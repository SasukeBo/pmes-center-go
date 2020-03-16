package ftpclient

// 编写读取ftp服务器数据的逻辑代码
// 解析CSV文件
// 生成批量插入的SQL

import (
	"bytes"
	"encoding/csv"
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
