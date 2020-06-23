package logic

// 访问ftp的task文件
// 注册ftp获取文件队列，worker
import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/errormap"
	"github.com/SasukeBo/ftpviewer/ftp"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/orm/types"
	timer "github.com/SasukeBo/lib/time"
	"github.com/SasukeBo/log"
	"github.com/tealeg/xlsx"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"
)

const fileNameDecodePattern = `([\w]*)-([\w]*)-.*-([A|B|w|b]?).xlsx`

type XLSXReader struct {
	DataSet        [][]string          // only cache data of sheet 1
	Material       *orm.Material       // 导入的料号
	Device         *orm.Device         // 导入设备
	Record         *orm.ImportRecord   // 导入记录
	DecodeTemplate *orm.DecodeTemplate // 解析模板
	Size           int                 // 文件大小
}

func newXLSXReader(material *orm.Material, device *orm.Device, template *orm.DecodeTemplate) *XLSXReader {
	return &XLSXReader{
		DataSet:        make([][]string, 0),
		Material:       material,
		Device:         device,
		DecodeTemplate: template,
	}
}

func (xr *XLSXReader) Read(path string) error {
	dataSheet, size, err := read(path)
	if err != nil {
		return err
	}
	var bIdx = xr.DecodeTemplate.DataRowIndex
	var eIdx = len(dataSheet) - 1
	for i, row := range dataSheet {
		if (len(row) == 0 || row[0] == "") && i > bIdx { // 空行 或 行首位空 为截至行，理想情况下不存在数据行中穿插空行
			eIdx = i - 1
			break
		}
	}
	dataSet := dataSheet[bIdx : eIdx+1]
	xr.DataSet = dataSet
	xr.Size = size

	log.Info("data begin idx: %v, end idx: %v\n", bIdx, eIdx)
	return nil
}

func read(path string) ([][]string, int, error) {
	content, err := ftp.ReadFile(path)
	size := len(content)
	if err != nil {
		if fe, ok := err.(*ftp.FTPError); ok {
			fe.Logger()
			return nil, 0, err
		}

		log.Errorln(err)
		return nil, 0, &ftp.FTPError{
			Message:   fmt.Sprintf("从FTP服务器读取文件%s失败", path),
			OriginErr: err,
		}
	}

	file, err := xlsx.OpenBinary(content)
	if err != nil {
		return nil, 0, fmt.Errorf("读取数据文件失败，原始错误信息: %v", err)
	}

	originData, err := file.ToSlice()
	if err != nil {
		return nil, 0, err
	}
	if len(originData) == 0 {
		return nil, 0, errors.New("xlsx文件内容是空的。")
	}

	return originData[0], size, nil
}

type fetchItem struct {
	Device   orm.Device
	FileName string
}

// FetchMaterialData 判断是否需要从FTP拉取数据
// 给定料号，对比数据库中已拉取文件路径，得出是否有需要拉取的文件路径
func FetchMaterialData(material *orm.Material) error {
	var needFetch []fetchItem

	template, err := material.GetDefaultTemplate()
	if err != nil {
		return errormap.NewOrigin("get default decode template for material(id = %v) failed: %v", material.ID, err)
	}

	ftpFileList, err := ftp.GetList("./" + material.Name)
	if err != nil {
		return err
	}

	for _, p := range ftpFileList {
		need, deviceRemark := checkFile(material.ID, p)
		if !need {
			continue
		}
		var device orm.Device
		device.CreateIfNotExist(material.ID, deviceRemark)
		needFetch = append(needFetch, fetchItem{device, p})
	}

	if len(needFetch) == 0 {
		return nil
	}

	return fetchMaterialData(*material, needFetch, template)
}

func fetchMaterialData(material orm.Material, files []fetchItem, dt *orm.DecodeTemplate) error {
	for _, f := range files {
		xr := newXLSXReader(&material, &f.Device, dt)
		path := resolvePath(material.Name, f.FileName)

		importRecord := &orm.ImportRecord{
			FileName:         filepath.Base(f.FileName),
			Path:             path,
			MaterialID:       material.ID,
			DeviceID:         f.Device.ID,
			Status:           orm.ImportStatusLoading,
			ImportType:       orm.ImportRecordTypeSystem,
			DecodeTemplateID: dt.ID,
		}
		if err := orm.Create(importRecord).Error; err != nil {
			// TODO: add log
			log.Errorln(err)
			continue
		}

		go func() {
			log.Warn("start read routine with file: %s\n", path)
			err := xr.Read(path)
			if err != nil {
				log.Error("read path(%s) error: %v", path, err)
				return
			}
			importRecord.RowCount = len(xr.DataSet)
			importRecord.FileSize = xr.Size
			if err := orm.Save(importRecord).Error; err != nil {
				// TODO: add log
				log.Errorln(err)
				return
			}
			xr.Record = importRecord
			go store(xr)
		}()
	}

	return nil
}

// checkFile 仅检查文件是否已经被读取到指定料号
func checkFile(materialID uint, fileName string) (bool, string) {
	var importRecord orm.ImportRecord
	// 查找 当前料号的 当前文件名的 已完成的 且 没有处理错误的 文件导入记录，若存在则忽略此文件
	orm.DB.Model(&importRecord).Where(
		"file_name = ? AND material_id = ? AND finished = 1 AND error IS NULL",
		fileName, materialID,
	).First(&importRecord)

	if importRecord.ID != 0 {
		return false, ""
	}

	reg := regexp.MustCompile(fileNameDecodePattern)
	matched := reg.FindStringSubmatch(fileName)
	if len(matched) != 4 {
		return false, ""
	}
	return true, matched[2]
}

func resolvePath(m, path string) string {
	return fmt.Sprintf("./%s/%s", m, filepath.Base(path))
}

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

// store xlsx data into db
func store(xr *XLSXReader) {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
			_ = xr.Record.Failed(errormap.ErrorCodeImportFailedWithPanic, err)
			debug.PrintStack()
		}
	}()

	var time1 = time.Now()

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where("material_id = ?", xr.Material.ID).Find(&points).Error; err != nil {
		_ = xr.Record.Failed(errormap.ErrorCodeImportGetPointsFailed, err)
		return
	}

	iColumns := xr.DecodeTemplate.ProductColumns["columns"]
	columns, ok := iColumns.([]interface{})
	if !ok {
		_ = xr.Record.Failed(errormap.ErrorCodeImportWithIllegalDecodeTemplate, "product columns stored in db is not a list")
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
	_ = xr.Record.Finish()

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

// AutoFetch 自动拉取
func AutoFetch() {
	log.Infoln("[AutoFetch] Begin auto fetch worker")
	fetch()
	for {
		select {
		case <-time.After(12 * time.Hour):
			fetch()
		}
	}
}

func fetch() {
	var materials []orm.Material
	err := orm.DB.Model(&orm.Material{}).Find(&materials).Error
	if err != nil {
		log.Error("[autoFetch] get materials error: %v", err)
	}

	for _, m := range materials {
		// TODO: add log
		log.Info("[autoFetch] fetch file(%s) data", m.Name)

		go func() {
			err := FetchMaterialData(&m)
			if err != nil {
				// TODO: add log
				log.Errorln(err)
			}
		}()
	}
}
