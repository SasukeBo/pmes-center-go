package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，worker
import (
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	timer "github.com/SasukeBo/lib/time"
	"github.com/SasukeBo/log"
	"runtime/debug"
	"strconv"
	"time"
)

var fetchQueue chan string
var cacheQueue chan *XLSXReader
var (
	singleInsertLimit = 10000
	insertProductsTpl = `
		INSERT INTO products (
			import_record_id,
			material_id,
			device_id,
			qualified,
			created_at,
			attribute,
			point_values
		)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?,?,?,?)`
	productValueCount    = 7
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

	var time1 = time.Now()

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("material_id = ?", xr.Material.ID).Find(&points).Error; err != nil {
		// TODO: add log
		return
	}

	iColumns := xr.DecodeTemplate.ProductColumns["columns"]
	columns, ok := iColumns.([]interface{})
	if !ok {
		// TODO: add log
		log.Error("decode template product columns error, got %+v\n", iColumns)
		return
	}

	productValueExpands := make([]interface{}, 0)

	for _, row := range xr.DataSet {
		qualified := true
		createdAt := timer.ParseTime(row[xr.DecodeTemplate.CreatedAtColumnIndex], 8)
		attribute := make(types.Map)
		for _, iColumn := range columns {
			column := iColumn.(map[string]interface{})
			index := int(column["Index"].(float64))
			cType := column["Type"].(string)
			name := column["Name"].(string)
			value := row[index]

			switch cType {
			case orm.ProductColumnTypeDatetime:
				t := timer.ParseTime(value, 8)
				if t == nil {
					now := time.Now()
					t = &now
				}

				attribute[name] = *t

			case orm.ProductColumnTypeFloat:
				fv, err := strconv.ParseFloat(value, 64)
				if err != nil {
					fv = float64(0)
				}
				attribute[name] = fv
			case orm.ProductColumnTypeInteger:
				iv, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					iv = int64(0)
				}
				attribute[name] = iv
			case orm.ProductColumnTypeString:
				attribute[name] = fmt.Sprint(value)
			}
		}

		pointValues := make(types.Map)
		for _, v := range points {
			ii, ok := xr.DecodeTemplate.PointColumns[v.Name]
			if !ok { // 模板中没有该名称点位的解析配置
				continue
			}
			idx := int(ii.(float64))
			value := parseFloat(row[idx])
			if value < v.LowerLimit || value > v.UpperLimit {
				qualified = false
			}
			pointValues[v.Name] = value
		}
		productValueExpands = append(productValueExpands,
			xr.Record.ID,
			xr.Material.ID,
			xr.Device.ID,
			qualified,
			*createdAt,
			attribute,
			pointValues,
		)
	}

	execInsert(productValueExpands, productValueCount, insertProductsTpl, productValueFieldTpl, xr.Record)
	xr.Record.Finished = true
	orm.Save(xr.Record)

	var time2 = time.Now()
	fmt.Printf("___________________________ process duration is %v\n", time2.Sub(time1))
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string, record *orm.ImportRecord) {
	tx := orm.DB.Begin()
	tx.LogMode(false)
	dataLen := len(dataset)
	totalLen := dataLen / itemLen
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
		record.RowFinishedCount = record.RowFinishedCount + singleInsertLimit
		orm.Save(record)
	}

	restLen := totalLen % singleInsertLimit
	if restLen > 0 {
		vSQL := valuetpl
		for j := 1; j < restLen; j++ {
			vSQL = vSQL + "," + valuetpl
		}
		end := dataLen
		begin := dataLen - restLen*itemLen
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert] %v\n", err)
		}
		record.RowFinishedCount = record.RowFinishedCount + restLen
		orm.Save(record)
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
}
