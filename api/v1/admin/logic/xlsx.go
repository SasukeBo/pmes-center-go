package logic

// 访问ftp的task文件
// 注册ftp获取文件队列，worker
import (
	"errors"
	"fmt"
	"github.com/SasukeBo/configer"
	timer "github.com/SasukeBo/lib/time"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/model"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/ftp"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

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

func (xr *XLSXReader) ReadFile(file *orm.File) error {
	content, err := ioutil.ReadFile(filepath.Join(configer.GetString("file_cache_path"), file.Path))
	if err != nil {
		return fmt.Errorf("读取文件失败：%v", err)
	}

	log.Info("file %s content length is %v", file.Name, len(content))

	return xr.setData(content)
}

func (xr *XLSXReader) ReadFTP(path string) error {
	content, err := ftp.ReadFile(path)
	if err != nil {
		if fe, ok := err.(*ftp.FTPError); ok {
			fe.Logger()
			return err
		}

		log.Errorln(err)
		return &ftp.FTPError{
			Message:   fmt.Sprintf("从FTP服务器读取文件%s失败", path),
			OriginErr: err,
		}
	}

	return xr.setData(content)
}

func (xr *XLSXReader) setData(content []byte) error {
	size := len(content)
	file, err := xlsx.OpenBinary(content)
	if err != nil {
		return fmt.Errorf("读取数据文件失败，原始错误信息: %v", err)
	}

	originData, err := file.ToSlice()
	if err != nil {
		return err
	}
	if len(originData) == 0 {
		return errors.New("xlsx文件内容是空的。")
	}

	dataSheet := originData[0]

	var bIdx = xr.DecodeTemplate.DataRowIndex
	var eIdx = len(dataSheet) - 1
	for i, row := range dataSheet {
		if (len(row) == 0 || row[0] == "") && i > bIdx { // 空行 或 行首位空 为截止行，理想情况下不存在数据行中穿插空行
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

type fetchItem struct {
	Device   orm.Device
	FileName string
}

// FetchMaterialData 判断是否需要从FTP拉取数据
// 给定料号，对比数据库中已拉取文件路径，得出是否有需要拉取的文件路径
func FetchMaterialData(material *orm.Material) error {
	var needFetch []string

	template, err := material.GetDefaultTemplate()
	if err != nil {
		return errormap.NewOrigin("get default decode template for material(id = %v) failed: %v", material.ID, err)
	}

	ftpFileList, err := ftp.GetList("./" + material.Name)
	if err != nil {
		return err
	}

	for _, p := range ftpFileList {
		need := checkFile(material.ID, p)
		if !need {
			continue
		}
		needFetch = append(needFetch, p)
	}

	if len(needFetch) == 0 {
		return nil
	}

	return fetchMaterialData(*material, needFetch, template)
}

func fetchMaterialData(material orm.Material, files []string, dt *orm.DecodeTemplate) error {
	for _, f := range files {
		xr := newXLSXReader(&material, nil, dt)
		path := resolvePath(material.Name, f)

		importRecord := &orm.ImportRecord{
			FileName:         filepath.Base(f),
			Path:             path,
			MaterialID:       material.ID,
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
			err := xr.ReadFTP(path)
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

// 直接根据数据库file记录获取数据
func FetchFileData(user orm.User, material orm.Material, device orm.Device, template orm.DecodeTemplate, tokens []string) error {
	var err error
	defer func() {
		errMessage := recover()
		if errMessage != nil {
			err = errors.New(fmt.Sprint(errMessage))
			debug.PrintStack()
		}
	}()

	for _, token := range tokens {
		xr := newXLSXReader(&material, &device, &template)
		var file orm.File
		if err := file.GetByToken(token); err != nil {
			log.Error("[FetchFileData] Get file with token=%s failed: %v", token, err)
			return err
		}

		if err := xr.ReadFile(&file); err != nil {
			message := fmt.Sprintf("[FetchFileData] Read file(%s) failed: %v", file.Name, err)
			log.Errorln(message)
			return errormap.NewCodeOrigin(errormap.ErrorCodeFileOpenFailedError, message)
		}

		importRecord := &orm.ImportRecord{
			FileID:           file.ID,
			FileName:         file.Name,
			Path:             file.Path,
			MaterialID:       material.ID,
			DeviceID:         device.ID,
			Status:           orm.ImportStatusLoading,
			ImportType:       orm.ImportRecordTypeUser,
			UserID:           user.ID,
			DecodeTemplateID: template.ID,
			RowCount:         len(xr.DataSet),
			FileSize:         xr.Size,
		}
		if err := orm.Create(importRecord).Error; err != nil {
			// TODO: add log
			log.Error("[FetchFileData] create import record failed: %v", err)
			continue
		}

		xr.Record = importRecord
		go store(xr)
	}

	return err
}

// checkFile 仅检查文件是否已经被读取到指定料号
func checkFile(materialID uint, fileName string) bool {
	var importRecord orm.ImportRecord
	// 查找 当前料号的 当前文件名的 已完成的 且 没有处理错误的 文件导入记录，若存在则忽略此文件
	orm.DB.Model(&importRecord).Where(
		"file_name = ? AND material_id = ? AND status = ?",
		fileName, materialID, model.ImportStatusFinished,
	).First(&importRecord)

	if importRecord.ID != 0 {
		return false
	}

	if !strings.Contains(fileName, ".xlsx") {
		return false
	}

	return true
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

	productColumns := xr.DecodeTemplate.ProductColumns
	productValueExpands := make([]interface{}, 0)

	for _, row := range xr.DataSet {
		qualified := true
		createdAt := timer.ParseTime(row[xr.DecodeTemplate.CreatedAtColumnIndex], 8)
		attribute := make(types.Map)
		for name, iColumn := range productColumns {
			column := iColumn.(map[string]interface{})
			index := int(column["Index"].(float64))
			cType := column["Type"].(string)
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
			if idx >= len(row) {
				message := fmt.Sprintf("point(%s) index(%d) out of range with data row length(%d)", v.Name, idx, len(row))
				_ = xr.Record.Failed(errormap.ErrorCodeImportWithIllegalDecodeTemplate, message)
				log.Errorln(message)
				return
			}
			value := parseFloat(row[idx])
			if value < v.LowerLimit || value > v.UpperLimit {
				qualified = false
			}
			pointValues[v.Name] = value
		}
		var deviceID uint
		if xr.Device != nil {
			deviceID = xr.Device.ID
		}
		productValueExpands = append(productValueExpands,
			xr.Record.ID,
			xr.Material.ID,
			deviceID,
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
