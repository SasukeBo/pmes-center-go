package ftpclient

// 访问ftp的task文件
// 注册ftp获取文件队列，woker
import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/SasukeBo/ftpviewer/orm"
)

var fetchQueue chan string
var cacheQueue chan *CSVDecoder

var (
	reg               *regexp.Regexp
	singleInsertLimit = 10000
	filenamePattern   = `(\d+)-(\d+)-(\d+)-[w|b]\.csv`
	timePattern       = `(\d{4})/(\d{2})/(\d{2}) (\d{2}:\d{2}:\d{2})`
	insertProductsTpl = `
		INSERT INTO products (product_uuid, material_id, device_id, qualified, producted_at)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?,?)`
	insertSizeValuesTpl  = `
		INSERT INTO size_values (size_id, product_uuid, size_values.value)
		VALUES
		%s
	`
	sizeValueFieldTpl = `(?,?,?)`
)

// FTPWorker _
func FTPWorker() {
	for {
		select {
		case path := <-fetchQueue:
			go fetchAndStore(path)
		case csvr := <-cacheQueue:
			go Store(csvr)
		}
	}
}

func fetchAndStore(path string) {
	csvr, err := Fetch(path)
	if err != nil {
		return
	}

	Store(csvr)
}

// Fetch fetch data from ftp server with file path
func Fetch(path string) (*CSVDecoder, error) {
	csvDecoder := CSVDecoder{}
	result := reg.FindAllStringSubmatch(filepath.Base(path), -1)
	if len(result) > 0 && len(result[0]) > 3 {
		csvDecoder.MaterialID = result[0][1]
		csvDecoder.DeviceName = fmt.Sprintf("%s设备%s", csvDecoder.MaterialID, result[0][2])
	} else {
		return nil, &FTPError{
			Message: fmt.Sprintf("文件名格式不正确，%s", path),
		}
	}

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

	csvDecoder.Decode([]byte(content))
	return &csvDecoder, nil
}

// Store csv data into db
func Store(csv *CSVDecoder) {
	mid := csv.MaterialID
	dn := csv.DeviceName
	rowLen := len(csv.Headers)
	sizeNames := csv.Headers[4 : rowLen-1]
	material := orm.GetMaterialWithIDCache(mid)
	if material == nil {
		material = &orm.Material{Name: mid}
		if err := orm.DB.Create(material).Error; err != nil {
			log.Println("create material failed, err: " + err.Error())
			return
		}
		orm.CacheMaterial(*material)
	}

	device := orm.GetDeviceWithNameCache(dn)
	if device == nil {
		device = &orm.Device{Name: dn, MaterialID: material.Name}
		if err := orm.DB.Create(device).Error; err != nil {
			log.Println("create device failed, err:" + err.Error())
			return
		}
		orm.CacheDevice(*device)
	}

	sizes := make([]orm.Size, 0)
	for i, sn := range sizeNames {
		size := orm.GetSizeWithMaterialIDSizeNameCache(sn, material.Name)
		if size == nil {
			upperLimit := parseFloat(csv.Limits[0][i+4])
			lowerLimit := parseFloat(csv.Limits[1][i+4])
			size = &orm.Size{
				Name:       sn,
				MaterialID: material.Name,
				UpperLimit: upperLimit,
				LowerLimit: lowerLimit,
			}
			if err := orm.DB.Create(size).Error; err != nil {
				log.Println("create size failed, err:" + err.Error())
				return
			}
			orm.CacheSize(*size)
			sizes = append(sizes, *size)
		}
	}

	products := make([]interface{}, 0)
	sizevalues := make([]interface{}, 0)
	for _, row := range csv.Rows {
		qualified := true
		puuid := row[rowLen-1]
		for j, v := range row[4 : rowLen-1] {
			value := parseFloat(v)
			size := sizes[j]
			sv := []interface{}{size.ID, puuid, value}
			if value < size.LowerLimit || value > size.UpperLimit {
				qualified = false
			}
			sizevalues = append(sizevalues, sv...)
		}

		productedAt := timeFormat(row[1])
		pv := []interface{}{puuid, material.Name, device.ID, qualified, productedAt}
		products = append(products, pv...)
	}

	execInsert(products, 5, insertProductsTpl, productValueFieldTpl)
	execInsert(sizevalues, 3, insertSizeValuesTpl, sizeValueFieldTpl)
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string) {
	datalen := len(dataset)
	totalLen := datalen / itemLen
	vSQL := valuetpl
	for i := 1; i < singleInsertLimit; i++ {
		vSQL = vSQL + "," + valuetpl
	}

	for i := 0; i < totalLen/singleInsertLimit; i++ {
		begin := i * singleInsertLimit * itemLen
		end := (i + 1) * singleInsertLimit * itemLen
		orm.DB.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...)
	}

	restLen := totalLen % singleInsertLimit
	if restLen > 0 {
		vSQL := valuetpl
		for j := 1; j < restLen; j++ {
			vSQL = vSQL + "," + valuetpl
		}
		end := datalen
		begin := datalen - restLen*itemLen
		orm.DB.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...)
	}
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
func PushStore(csvr *CSVDecoder) {
	cacheQueue <- csvr
}

func init() {
	fetchQueue = make(chan string, 10)
	cacheQueue = make(chan *CSVDecoder, 10)
	reg = regexp.MustCompile(filenamePattern)
}
