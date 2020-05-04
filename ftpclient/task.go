package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，woker
import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"regexp"
	"strconv"
)

var fetchQueue chan string
var cacheQueue chan *XLSXReader

var (
	reg               *regexp.Regexp
	singleInsertLimit = 10000
	filenamePattern   = `(.*)-(.*)-(.*)-([w|b])\.xlsx`
	timePattern       = `(\d{4})/(\d{2})/(\d{2}) (\d{2}:\d{2}:\d{2})`
	insertProductsTpl = `
		INSERT INTO products (product_uuid, material_id, device_id, qualified, created_at)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?,?)`
	insertSizeValuesTpl  = `
		INSERT INTO size_values (device_id, size_id, product_uuid, size_values.value, qualified, created_at)
		VALUES
		%s
	`
	sizeValueFieldTpl = `(?,?,?,?,?,?)`
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
			log.Println(err)
		}
	}()
	materialName := xr.MaterialID
	deviceName := xr.DeviceName
	material := orm.GetMaterialWithName(materialName)
	if material == nil {
		return
	}

	device := orm.GetDeviceWithName(deviceName)
	if device == nil {
		return
	}

	var sizes []orm.Size
	if err := orm.DB.Where("material_id = ?", material.ID).Find(&sizes).Error; err != nil {
		fmt.Println(err)
		return
	}

	products := make([]interface{}, 0)
	sizeValues := make([]interface{}, 0)
	for i, row := range xr.DateSet {
		if row[0] == "" { // 过滤掉空行
			continue
		}
		qp := true
		puuid := fmt.Sprintf("%s%v", xr.ProductUUIDPrefix, i)
		for _, v := range sizes {
			qs := true
			value := parseFloat(row[v.Index])
			if value < v.LowerLimit || value > v.UpperLimit {
				qp = false
				qs = false
			}
			sv := []interface{}{device.ID, v.ID, puuid, value, qs, *xr.ProductAt}
			sizeValues = append(sizeValues, sv...)
		}

		pv := []interface{}{puuid, material.ID, device.ID, qp, *xr.ProductAt}
		products = append(products, pv...)
	}

	fmt.Println("-----------------------------------\nbegin execInsert ....")
	execInsert(products, 5, insertProductsTpl, productValueFieldTpl)
	execInsert(sizeValues, 6, insertSizeValuesTpl, sizeValueFieldTpl)
	fmt.Println("-----------------------------------\nfinish execInsert ....")
	orm.DB.Model(&orm.FileList{}).Where("id = ?", xr.PathID).Update("finished", true)
	fmt.Printf("-----------------------------------\nfinish udpate file id=%v finished=true\n", xr.PathID)
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
		if err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error; err != nil {
			log.Println(err)
		}
	}
	tx.Commit()
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
