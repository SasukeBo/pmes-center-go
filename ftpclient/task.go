package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，woker
import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	stime "github.com/SasukeBo/lib/time"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	productValueCount = 10
	insertPointValuesTpl = `
		INSERT INTO point_values (point_id, product_uuid, v)
		VALUES
		%s
	`
	pointValueFieldTpl = `(?,?,?)`
	pointValueCount = 3
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
	materialName := xr.MaterialID
	deviceName := xr.DeviceName
	material := orm.GetMaterialWithName(materialName)
	if material == nil {
		fmt.Printf("material %s not found", materialName)
		return
	}

	device := orm.GetDeviceWithName(deviceName)
	if device == nil {
		fmt.Printf("device %s not found", deviceName)
		return
	}

	var sizeIDs []int
	if err := orm.DB.Model(&orm.Size{}).Where("material_id = ?", material.ID).Pluck("id", &sizeIDs).Error; err != nil {
		fmt.Println(err)
		return
	}

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("size_id in (?)", sizeIDs).Find(&points).Error; err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("-----------------------------------\nhandle rows ....")
	products := make([]interface{}, 0)
	pointValues := make([]interface{}, 0)
	for _, row := range xr.DateSet {
		if !validRow(row) { // 过滤掉无效行和空行
			continue
		}
		qp := true
		// 生产 product uuid
		puuid := uuid.New().String()
		productAt, err := stime.ParseTime(row[1], 8)
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

		lineID, _ := strconv.Atoi(row[3])
		mouldID, _ := strconv.Atoi(row[5])
		shiftNumber, _ := strconv.Atoi(row[6])

		pv := []interface{}{puuid, material.ID, device.ID, qp, productAt, row[2], lineID, row[4], mouldID, shiftNumber}
		products = append(products, pv...)
	}

	total := (len(products) / productValueCount) + (len(pointValues)/pointValueCount)
	orm.DB.Model(&orm.File{}).Where("id = ?", xr.PathID).Update("total_rows", total)

	fmt.Println("-----------------------------------\nbegin execInsert ....")
	execInsert(products, productValueCount, insertProductsTpl, productValueFieldTpl, xr.PathID)
	execInsert(pointValues, pointValueCount, insertPointValuesTpl, pointValueFieldTpl, xr.PathID)
	fmt.Println("-----------------------------------\nfinish execInsert ....")
}

func validRow(row []string) bool {
	if len(row) == 0 {
		fmt.Println("empty row")
		return false
	}
	if row[0] == "" {
		return false
	}

	if _, err := strconv.Atoi(row[0]); err != nil {
		return false
	}

	return true
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string, fileID int) {
	tx := orm.DB.Begin()
	tx.LogMode(false)
	datalen := len(dataset)
	totalLen := datalen / itemLen
	vSQL := valuetpl
	for i := 1; i < singleInsertLimit; i++ {
		vSQL = vSQL + "," + valuetpl
	}

	var file orm.File
	orm.DB.Model(&file).Where("id = ?", fileID).First(&file)
	finishedRows := file.FinishedRows
	for i := 0; i < totalLen/singleInsertLimit; i++ {
		begin := i * singleInsertLimit * itemLen
		end := (i + 1) * singleInsertLimit * itemLen
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert] %v\n", err)
		}
		finishedRows = finishedRows + singleInsertLimit
		updateFinishedRows(fileID, finishedRows)
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
		finishedRows = finishedRows + restLen
		updateFinishedRows(fileID, finishedRows)
	}
	// 最后完成该文件
	orm.DB.Model(&orm.File{}).Where("id = ?", fileID).Update("finished", true)
	tx.Commit()
}

func updateFinishedRows(fileID, finishedRows int) {
	orm.DB.Model(&orm.File{}).Where("id = ?", fileID).Update("finished_rows", finishedRows)
	fmt.Printf("-----------------------------------\nfinish udpate file id=%v finished rows=%v\n", fileID, finishedRows)
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
