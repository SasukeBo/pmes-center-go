package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，woker
import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
)

var fetchQueue chan string
var cacheQueue chan *XLSXReader

var (
	reg               *regexp.Regexp
	singleInsertLimit = 10000
	filenamePattern   = `(\d+)-(\d+)-(\d+)-[w|b]\.xlsx`
	timePattern       = `(\d{4})/(\d{2})/(\d{2}) (\d{2}:\d{2}:\d{2})`
	insertProductsTpl = `
		INSERT INTO products (product_uuid, material_id, device_id, qualified)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?)`
	insertSizeValuesTpl  = `
		INSERT INTO size_values (size_name, product_uuid, size_values.value, qualified)
		VALUES
		%s
	`
	sizeValueFieldTpl = `(?,?,?,?)`
)

// FTPWorker _
func FTPWorker() {
	for {
		select {
		case path := <-fetchQueue:
			go fetchAndStore(path)
		case xr := <-cacheQueue:
			go Store(xr)
		}
	}
}

func fetchAndStore(path string) {
	xr := NewXLSXReader()
	if err := xr.Read(path); err != nil {
		log.Println(err)
		return
	}

	Store(xr)
}

// Store xlsx data into db
func Store(xr *XLSXReader) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	mid := xr.MaterialID
	dn := xr.DeviceName
	material := orm.GetMaterialWithID(mid)
	if material == nil {
		return
	}

	device := orm.GetDeviceWithName(dn)
	if device == nil {
		return
	}

	products := make([]interface{}, 0)
	sizeValues := make([]interface{}, 0)
	for i, row := range xr.DateSet {
		qp := true
		puuid := fmt.Sprintf("%s%v", xr.ProductUUIDPrefix, i)
		for k, v := range xr.DimSL {
			qs := true
			value := parseFloat(row[v.Index])
			if value < v.LSL || value > v.USL {
				qp = false
				qs = false
			}
			sv := []interface{}{k, puuid, value, qs}
			sizeValues = append(sizeValues, sv...)
		}

		pv := []interface{}{puuid, material.Name, device.ID, qp}
		products = append(products, pv...)
	}

	execInsert(products, 4, insertProductsTpl, productValueFieldTpl)
	execInsert(sizeValues, 4, insertSizeValuesTpl, sizeValueFieldTpl)
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string) {
	tx := orm.DB.Begin()
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
		tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...)
	}

	restLen := totalLen % singleInsertLimit
	if restLen > 0 {
		vSQL := valuetpl
		for j := 1; j < restLen; j++ {
			vSQL = vSQL + "," + valuetpl
		}
		end := datalen
		begin := datalen - restLen*itemLen
		tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...)
	}
	tx.Commit()
}

func timeFormat(t string) time.Time {
	r := regexp.MustCompile(timePattern)
	re := r.FindAllStringSubmatch(t, -1)
	if len(re) > 0 && len(re[0]) == 5 {
		match := re[0]
		timeStr := fmt.Sprintf("%s-%s-%sT%s+08:00", match[1], match[2], match[3], match[4])
		t, _ := time.Parse(time.RFC3339, timeStr)
		return t
	}

	return time.Now()
}

func parseFloat(v string) float64 {
	fv, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return fv
}

// PushFetch _
func PushFetch(path string) {
	fetchQueue <- path
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
