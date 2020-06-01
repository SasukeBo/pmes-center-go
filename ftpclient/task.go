package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，worker
import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	timer "github.com/SasukeBo/lib/time"
	"github.com/SasukeBo/log"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"
)

var fetchQueue chan string
var cacheQueue chan *XLSXReader

var (
	reg               *regexp.Regexp
	singleInsertLimit = 10000
	filenamePattern   = `(.*)-(.*)-(.*)-([w|b])\.xlsx`
	insertProductsTpl = `
		INSERT INTO products (uuid, material_id, device_id, qualified, created_at, d2_code, line_id, jig_id, mould_id, shift_number)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?,?,?,?,?,?,?)`
	productValueCount    = 10
	insertPointValuesTpl = `
		INSERT INTO point_values (point_id, product_uuid, v)
		VALUES
		%s
	`
	pointValueFieldTpl = `(?,?,?)`
	pointValueCount    = 3
)

// FTPWorker _
func FTPWorker() {
	for {
		select {
		case xr := <-cacheQueue:
			fmt.Println("--------------------------\nstart store task")
			go Store(xr)
		}
	}
}

// Store xlsx data into db
func Store(xr *XLSXReader) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			debug.PrintStack()
		}
	}()

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("material_id = ?", xr.Material.ID).Find(&points).Error; err != nil {
		// TODO: add log
		return
	}

	columns, ok := xr.DecodeTemplate.ProductColumns["columns"].([]orm.Column)
	if !ok {
		// TODO: add log
		log.Errorln("decode template product columns error")
		return
	}

	products := make([]interface{}, 0)
	pointValues := make([]interface{}, 0)
	for _, row := range xr.DataSet {
		if !validRow(row) { // 过滤掉无效行和空行
			continue
		}
		var qp = true
		for _, column := range columns {
			// TODO: parse number
			switch column.Type {
			case orm.ProductColumnTypeDatetime:

			}
		}

		// TODO: time zone config
		// TODO: 在本地实现时间解析
		productAt, err := timer.ParseTime(row[1], 8)
		if err != nil {
			t := time.Now()
			productAt = &t
		}


		for _, v := range points {
			value := parseFloat(row[v.Index])
			if value < v.LowerLimit || value > v.UpperLimit {
				qp = false
			}
			pv := []interface{}{v.ID, puuid, value}
			pointValues = append(pointValues, pv...)
		}

		pv := []interface{}{puuid, material.ID, device.ID, qp, productAt, row[2], row[3], row[4], row[5], row[6]}
		products = append(products, pv...)
	}

	finishChan := make(chan int, 0)
	go execInsert(products, productValueCount, insertProductsTpl, productValueFieldTpl, xr.PathID, finishChan)
	go execInsert(pointValues, pointValueCount, insertPointValuesTpl, pointValueFieldTpl, xr.PathID, finishChan)

	f := 0
	for {
		c := <-finishChan
		f = f + c
		if f == 2 {
			break
		}
	}

	// 最后完成该文件
	orm.DB.Model(&orm.File{}).Where("id = ?", xr.PathID).Update("finished", true)
}

func validRow(row []string) bool {
	if len(row) == 0 { // 过滤空行
		return false
	}
	if row[0] == "" { // 过滤第一个单元格空的行
		return false
	}

	if _, err := strconv.Atoi(row[0]); err != nil { // 过滤首单元格不可转数字的行
		return false
	}

	return true
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string, fileID int, finishChan chan int) {
	tx := orm.DB.Begin()
	//tx.LogMode(true)
	tx.LogMode(false)
	datalen := len(dataset)
	totalLen := datalen / itemLen
	vSQL := valuetpl
	for i := 1; i < singleInsertLimit; i++ {
		vSQL = vSQL + "," + valuetpl
	}

	for i := 0; i < totalLen/singleInsertLimit; i++ {
		begin := i * singleInsertLimit * itemLen
		end := (i + 1) * singleInsertLimit * itemLen
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert] %v\n", err)
		}
		updateFinishedRows(fileID, singleInsertLimit)
	}

	restLen := totalLen % singleInsertLimit
	if restLen > 0 {
		vSQL := valuetpl
		for j := 1; j < restLen; j++ {
			vSQL = vSQL + "," + valuetpl
		}
		end := datalen
		begin := datalen - restLen*itemLen
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert] %v\n", err)
		}
		updateFinishedRows(fileID, restLen)
	}

	tx.Commit()
	finishChan <- 1
}

func updateFinishedRows(fileID, plus int) {
	//var file orm.File
	//orm.DB.Model(&file).Where("id = ?", fileID).First(&file)
	//orm.DB.Model(&orm.File{}).Where("id = ?", fileID).Update("finished_rows", file.FinishedRows+plus)
	//// fmt.Printf("-----------------------------------\nfinish udpate file id=%v finished rows=%v\n", fileID, file.FinishedRows+plus)
}

func parseFloat(v string) float64 {
	fv, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return fv
}

// PushStore _
func PushStore(xr *XLSXReader) {
	cacheQueue <- xr
}

func init() {
	fetchQueue = make(chan string, 10)
	cacheQueue = make(chan *XLSXReader, 10)
	reg = regexp.MustCompile(filenamePattern)
}
