package logic

// 访问ftp的task文件
// 注册ftp获取文件队列，worker
import (
	"errors"
	"fmt"
	"github.com/SasukeBo/configer"
	timer "github.com/SasukeBo/lib/time"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/ftp"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/orm/types"
	"github.com/google/uuid"
	"github.com/tealeg/xlsx/v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

const (
	fileModeReadWrite os.FileMode = 0666
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
	baseDir := configer.GetString("file_cache_path")
	content, err := ioutil.ReadFile(filepath.Join(baseDir, file.Path))
	if err != nil {
		return fmt.Errorf("读取文件失败：%v", err)
	}

	log.Info("file %s content length is %v", file.Name, len(content))

	return xr.setData(content)
}

// 处理文件关联设备
func handleFileDevice(xr *XLSXReader, file *xlsx.File) error {
	if len(file.Sheets) == 0 {
		return fmt.Errorf("file has no sheet")
	}
	sheet := file.Sheets[0]
	firstRow, err := sheet.Row(0)
	if err != nil {
		return fmt.Errorf("first sheet of the file is empty")
	}
	firstCell := firstRow.GetCell(0)
	var deviceName string
	if firstCell != nil {
		deviceName = firstCell.String()
	}

	var device orm.Device
	if deviceName == "" {
		device = xr.Material.GetUnKnownDevice()
	} else {
		err = orm.Model(&device).Where("material_id = ? AND name = ?", xr.Material.ID, deviceName).First(&device).Error
		if err != nil {
			device = orm.Device{
				Name:       deviceName,
				Remark:     strings.ToLower(strings.ReplaceAll(deviceName, " ", "_")),
				MaterialID: xr.Material.ID,
			}
			orm.Save(&device)
		}
	}

	xr.Device = &device
	return nil
}

// ReadFTP 从FTP服务器读取文件
// 删除已读取的文件
// 存储文件到本地缓存
func (xr *XLSXReader) ReadFTP(path string) error {
	// 读取文件
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
	// 获取配置
	dst := configer.GetString("file_cache_path")

	// 创建目录
	var relevantPath = filepath.Join(orm.DirSource, xr.Material.Name, "default")
	var directory = filepath.Join(dst, relevantPath)
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return &ftp.FTPError{
			Message:   fmt.Sprintf("create directory(%s) failed: %v", directory, err),
			OriginErr: err,
		}
	}

	// 存储文件到本地
	token, err := uuid.NewRandom()
	if err != nil {
		return &ftp.FTPError{
			Message:   fmt.Sprintf("failed to generate token for file: %v", err),
			OriginErr: err,
		}
	}
	var localPath = filepath.Join(directory, token.String())
	if err := ioutil.WriteFile(localPath, content, fileModeReadWrite); err != nil {
		return &ftp.FTPError{
			Message:   fmt.Sprintf("save file to localpath(%s) failed: %v", localPath, err),
			OriginErr: err,
		}
	}
	var file = &orm.File{
		Name:        filepath.Base(path),
		Path:        filepath.Join(relevantPath, token.String()),
		Token:       token.String(),
		Size:        uint(len(content)),
		ContentType: orm.XlsxContentType,
	}
	orm.Create(file)

	// 删除Ftp文件
	_ = ftp.RemoveFile(path)

	xr.Record.FileID = file.ID
	if err := xr.setData(content); err != nil {
		return err
	}

	return nil
}

func (xr *XLSXReader) setData(content []byte) error {
	size := len(content)
	file, err := xlsx.OpenBinary(content) // TODO: 该步骤很慢
	if err != nil {
		return fmt.Errorf("读取数据文件失败，原始错误信息: %v", err)
	}
	formatTimeOfXlsx(xr.DecodeTemplate, file)
	if xr.Device == nil {
		if err := handleFileDevice(xr, file); err != nil {
			return err
		}
	}

	originData, err := file.ToSlice()
	if err != nil {
		log.Error("[setData] file.ToSlice(): %v", err)
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
	// update record
	xr.Record.FileSize = size
	xr.Record.RowCount = len(dataSet)
	orm.Save(xr.Record)

	log.Info("data begin idx: %v, end idx: %v\n", bIdx, eIdx)
	return nil
}

const xlsxDateFormatForNumeric = "yyyy/mm/dd hh:mm:ss"

func formatTimeOfXlsx(template *orm.DecodeTemplate, file *xlsx.File) error {
	idx := template.CreatedAtColumnIndex - 1
	beginRow := template.DataRowIndex - 1
	if len(file.Sheets) == 0 {
		return errors.New("file has no sheet")
	}
	sheet := file.Sheets[0]
	var i int
	sheet.ForEachRow(func(r *xlsx.Row) error {
		defer func() { i++ }()
		if i < beginRow {
			cell := r.GetCell(idx)
			cell.SetDateTimeWithFormat(0, xlsxDateFormatForNumeric)
			return nil
		}
		if cell := r.GetCell(idx); cell.Type() == xlsx.CellTypeNumeric {
			v, err := cell.Float()
			if err == nil {
				cell.SetDateTimeWithFormat(v, xlsxDateFormatForNumeric)
				fmt.Println(cell.String())
			}
		}

		return nil
	})

	return nil
}

// FetchMaterialData
// 给定料号，拉取Ftp服务器料号数据
func FetchMaterialData(material *orm.Material) error {
	template, err := material.GetCurrentVersionTemplate()
	if err != nil {
		return errormap.NewOrigin("get default decode template for material(id = %v) failed: %v", material.ID, err)
	}

	fetchList, err := ftp.GetDeepFilePath("./" + material.Name)
	if err != nil {
		return err
	}

	if len(fetchList) == 0 {
		return nil
	}

	return fetchMaterialData(*material, fetchList, template)
}

func fetchMaterialData(material orm.Material, paths []string, dt *orm.DecodeTemplate) error {
	for _, path := range paths {
		xr := newXLSXReader(&material, nil, dt)

		importRecord := &orm.ImportRecord{
			FileName:          filepath.Base(path),
			MaterialID:        material.ID,
			Status:            orm.ImportStatusLoading,
			ImportType:        orm.ImportRecordTypeSystem,
			DecodeTemplateID:  dt.ID,
			MaterialVersionID: dt.MaterialVersionID,
		}
		if err := orm.Create(importRecord).Error; err != nil {
			log.Errorln(err)
			continue
		}
		xr.Record = importRecord

		go func() {
			log.Warn("start read routine with file: %s\n", path)
			err := xr.ReadFTP(path)
			if err != nil {
				log.Error("read path(%s) error: %v", path, err)
				return
			}
			go store(xr)
		}()
	}

	return nil
}

// 直接根据数据库file记录获取数据
func FetchFileData(user orm.User, material orm.Material, device orm.Device, tokens []string) error {
	var err error
	defer func() {
		errMessage := recover()
		if errMessage != nil {
			err = errors.New(fmt.Sprint(errMessage))
			debug.PrintStack()
		}
	}()

	template, err := material.GetCurrentVersionTemplate()
	if err != nil {
		return err
	}

	for _, token := range tokens {
		var file orm.File
		if err := file.GetByToken(token); err != nil {
			log.Error("[FetchFileData] Get file with token=%s failed: %v", token, err)
			return err
		}
		xr := newXLSXReader(&material, &device, template)
		importRecord := &orm.ImportRecord{
			FileID:            file.ID,
			FileName:          file.Name,
			Path:              file.Path,
			MaterialID:        material.ID,
			DeviceID:          device.ID,
			Status:            orm.ImportStatusLoading,
			ImportType:        orm.ImportRecordTypeUser,
			UserID:            user.ID,
			MaterialVersionID: template.MaterialVersionID,
			FileSize:          xr.Size,
		}
		if err := orm.Create(importRecord).Error; err != nil {
			// TODO: add log
			log.Error("[FetchFileData] create import record failed: %v", err)
			continue
		}
		xr.Record = importRecord

		go func() {
			if err := xr.ReadFile(&file); err != nil {
				err := fmt.Errorf("[FetchFileData] Read file(%s) failed: %v", file.Name, err)
				xr.Record.Failed(errormap.ErrorCodeFileOpenFailedError, err)
				return
			}

			xr.Record.Status = orm.ImportStatusImporting
			orm.Save(xr.Record)

			go store(xr)
		}()
	}

	return err
}

//// checkFile 仅检查文件是否已经被读取到指定料号
// TODO: 遗弃
//func checkFile(materialID uint, fileName string) bool {
//	var importRecord orm.ImportRecord
//	// 查找 当前料号的 当前文件名的 已完成的 且 没有处理错误的 文件导入记录，若存在则忽略此文件
//	orm.DB.Model(&importRecord).Where(
//		"file_name = ? AND material_id = ? AND status = ?",
//		fileName, materialID, model.ImportStatusFinished,
//	).First(&importRecord)
//
//	if importRecord.ID != 0 {
//		return false
//	}
//
//	if !strings.Contains(fileName, ".xlsx") {
//		return false
//	}
//
//	return true
//}

//func resolvePath(m, path string) string {
//	return fmt.Sprintf("./%s/%s", m, filepath.Base(path))
//}

var (
	singleInsertLimit = 5000
	insertProductsTpl = `
		INSERT INTO products (
			import_record_id,
			material_id,
			device_id,
			qualified,
			created_at,
			attribute,
			point_values,
			material_version_id,
			bar_code_status,
			bar_code
		)
		VALUES
		%s
	`
	productValueFieldTpl = `(?,?,?,?,?,?,?,?,?,?)`
	productValueCount    = 10
)

// TODO deprecated
//func store(xr *XLSXReader) {
//	defer func() {
//		err := recover()
//		if err != nil {
//			fmt.Println(err)
//			_ = xr.Record.Failed(errormap.ErrorCodeImportFailedWithPanic, err)
//			debug.PrintStack()
//		}
//	}()
//
//	var time1 = time.Now()
//
//	var versions []orm.MaterialVersion
//	if err := orm.Model(&orm.MaterialVersion{}).Where("material_id = ? AND active = true", xr.Material.ID).Find(&versions).Error; err != nil {
//		_ = xr.Record.Failed(errormap.ErrorCodeActiveVersionNotFound, err)
//		return
//	}
//	if len(versions) == 0 {
//		_ = xr.Record.Failed(errormap.ErrorCodeActiveVersionNotFound, nil)
//		return
//	}
//	if len(versions) > 1 {
//		_ = xr.Record.Failed(errormap.ErrorCodeActiveVersionNotUnique, nil)
//		return
//	}
//
//	var currentVersion = versions[0]
//	var points []orm.Point
//	if err := orm.DB.Model(&orm.Point{}).Where(
//		"material_id = ? AND material_version_id = ?", xr.Material.ID, currentVersion.ID,
//	).Find(&points).Error; err != nil {
//		_ = xr.Record.Failed(errormap.ErrorCodeImportGetPointsFailed, err)
//		return
//	}
//
//	productColumns := xr.DecodeTemplate.ProductColumns
//	productValueExpands := make([]interface{}, 0)
//
//	var importOK int
//	for _, row := range xr.DataSet {
//		qualified := true
//		createdAt := time.Now()
//		if t := timer.ParseTime(row[xr.DecodeTemplate.CreatedAtColumnIndex-1], 8); t != nil {
//			createdAt = *t
//		}
//		attribute := make(types.Map)
//		for name, iColumn := range productColumns {
//			column := iColumn.(map[string]interface{})
//			index := int(column["Index"].(float64))
//			cType := column["Type"].(string)
//			value := row[index-1]
//
//			switch cType {
//			case orm.ProductColumnTypeDatetime:
//				t := timer.ParseTime(value, 8)
//				if t == nil {
//					now := time.Now()
//					t = &now
//				}
//				attribute[name] = *t
//			case orm.ProductColumnTypeFloat:
//				fv, err := strconv.ParseFloat(value, 64)
//				if err != nil {
//					fv = float64(0)
//				}
//				attribute[name] = fv
//			case orm.ProductColumnTypeInteger:
//				iv, err := strconv.ParseInt(value, 10, 64)
//				if err != nil {
//					iv = int64(0)
//				}
//				attribute[name] = iv
//			case orm.ProductColumnTypeString:
//				attribute[name] = fmt.Sprint(value)
//			}
//		}
//
//		pointValues := make(types.Map)
//		for _, v := range points {
//			idx := v.Index - 1
//			if idx >= len(row) {
//				message := fmt.Sprintf("point(%s) index(%d) out of range with data row length(%d)", v.Name, idx, len(row))
//				_ = xr.Record.Failed(errormap.ErrorCodeImportWithIllegalDecodeTemplate, message)
//				log.Errorln(message)
//				return
//			}
//			value := parseFloat(row[idx])
//			if value < v.LowerLimit || value > v.UpperLimit {
//				qualified = false
//			}
//			pointValues[v.Name] = value
//		}
//		var deviceID uint
//		if xr.Device != nil {
//			deviceID = xr.Device.ID
//		}
//
//		productValueExpands = append(productValueExpands,
//			xr.Record.ID,
//			xr.Material.ID,
//			deviceID,
//			qualified,
//			createdAt,
//			attribute,
//			pointValues,
//			currentVersion.ID,
//		)
//		if qualified {
//			importOK++
//		}
//	}
//
//	execInsert(productValueExpands, productValueCount, insertProductsTpl, productValueFieldTpl, xr.Record)
//
//	/*			记录单次导入良率
//	---------------------------------------------------------------------------------------------------------------- */
//	var yield float64
//	if total := len(xr.DataSet); total == 0 {
//		yield = 0
//	} else {
//		yield = float64(importOK) / float64(total)
//	}
//	_ = xr.Record.Finish(yield)
//
//	/*			记录当前版本的总量与良率
//	---------------------------------------------------------------------------------------------------------------- */
//	if err := currentVersion.UpdateWithRecord(xr.Record); err != nil {
//		log.Error("[store] currentVersion update with record failed: %v", err)
//	}
//
//	var time2 = time.Now()
//	fmt.Printf("___________________________ process file [%s] duration is %v\n", xr.Record.FileName, time2.Sub(time1))
//}

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

	currentVersion, err := xr.Material.GetCurrentVersion()
	if err != nil {
		_ = xr.Record.Failed(errormap.ErrorCodeActiveVersionNotFound, err)
		return
	}

	var points []orm.Point
	if err := orm.DB.Model(&orm.Point{}).Where(
		"material_id = ? AND material_version_id = ?", xr.Material.ID, currentVersion.ID,
	).Find(&points).Error; err != nil {
		_ = xr.Record.Failed(errormap.ErrorCodeImportGetPointsFailed, err)
		return
	}

	var decoder *BarCodeDecoder
	rule := xr.Material.GetCurrentTemplateDecodeRule()
	if rule != nil {
		decoder = NewBarCodeDecoder(rule)
	}

	productValueExpands := make([]interface{}, 0)

	var importOK, invalidRow int
	for _, row := range xr.DataSet {
		qualified := true
		rowValid := true
		createdAt := time.Now()
		if t := timer.ParseTime(row[xr.DecodeTemplate.CreatedAtColumnIndex-1], 8); t != nil {
			createdAt = *t
		}

		var attribute types.Map
		var statusCode = 1

		barCode := row[xr.DecodeTemplate.BarCodeIndex-1]
		barCode = strings.TrimSpace(barCode)
		if decoder != nil {
			attribute, statusCode = decoder.Decode(barCode)
		} else {
			attribute = make(types.Map)
			statusCode = orm.BarCodeStatusNoRule
		}

		pointValues := make(types.Map)
		for _, v := range points {
			idx := v.Index - 1
			if idx >= len(row) {
				message := fmt.Sprintf("point(%s) index(%d) out of range with data row length(%d)", v.Name, idx, len(row))
				_ = xr.Record.Failed(errormap.ErrorCodeImportWithIllegalDecodeTemplate, message)
				log.Errorln(message)
				return
			}
			var value float64
			value, rowValid = v.ValueWithLegal(row[idx])
			if !rowValid { // 无效数据结束遍历该行
				invalidRow++
				break
			}

			if value < v.LowerLimit || value > v.UpperLimit {
				qualified = false
			}
			pointValues[v.Name] = value
		}

		if !rowValid {
			continue // 过滤无效行
		}

		// TODO: fix device id
		var deviceID uint
		if xr.Device != nil {
			deviceID = xr.Device.ID
		}

		productValueExpands = append(productValueExpands,
			xr.Record.ID,
			xr.Material.ID,
			deviceID,
			qualified,
			createdAt,
			attribute,
			pointValues,
			currentVersion.ID,
			statusCode,
			barCode,
		)
		if qualified {
			importOK++
		}
	}

	execInsert(productValueExpands, productValueCount, insertProductsTpl, productValueFieldTpl, xr.Record)

	/*			记录单次导入良率
	---------------------------------------------------------------------------------------------------------------- */
	var yield float64
	if total := len(xr.DataSet); total == 0 {
		yield = 0
	} else {
		yield = float64(importOK) / float64(total)
	}
	xr.Record.RowInvalidCount = invalidRow
	_ = xr.Record.Finish(yield)

	/*			记录当前版本的总量与良率
	---------------------------------------------------------------------------------------------------------------- */
	if err := currentVersion.UpdateWithRecord(xr.Record); err != nil {
		log.Error("[store] currentVersion update with record failed: %v", err)
	}

	var time2 = time.Now()
	log.Info("___________________________ process file [%s] duration is %v\n", xr.Record.FileName, time2.Sub(time1))
}

func execInsert(dataset []interface{}, itemLen int, sqltpl, valuetpl string, record *orm.ImportRecord) {
	fmt.Printf("dataset length: %v\nitemLen: %v\n", len(dataset), itemLen)

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
		fmt.Printf("dataset[begin:end] length: %v\n", len(dataset[begin:end]))
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert LoopInsert] %v\n", err)
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
		fmt.Printf("dataset[begin:end] length: %v\n", len(dataset[begin:end]))
		err := tx.Exec(fmt.Sprintf(sqltpl, vSQL), dataset[begin:end]...).Error
		if err != nil {
			fmt.Printf("[execInsert restInsert] %v\n", err)
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
		case <-time.After(1 * time.Hour):
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
		log.Info("[autoFetch] fetch data for %s", m.Name)

		go func() {
			err := FetchMaterialData(&m)
			if err != nil {
				// TODO: add log
				log.Errorln(err)
			}
		}()
	}
}
