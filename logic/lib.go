package logic

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"github.com/jinzhu/gorm"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SasukeBo/ftpviewer/ftpclient"
)

var fileNamePattern = `(.*)-(.*)-(.*)-([w|b])\.xlsx`

// AutoFetch 自动拉取
func AutoFetch() {
	log.Infoln("[AutoFetch] Begin auto fetch worker")
	autoFetch()
	for {
		select {
		case <-time.After(12 * time.Hour):
			autoFetch()
		}
	}
}

func autoFetch() {
	var materials []orm.Material
	err := orm.DB.Model(&orm.Material{}).Find(&materials).Error
	if err != nil {
		log.Error("[autoFetch] get materials error: %v", err)
	}

	config := orm.GetSystemConfig("cache_days")
	cacheDays, err := strconv.Atoi(config.Value)
	if err != nil {
		cacheDays = 30
	}
	now := time.Now()
	begin := now.AddDate(0, 0, -cacheDays)

	for _, m := range materials {
		log.Info("[autoFetch] fetch %s duration %v - %v...", m.Name, begin, now)
		go func() {
			fileIDs, err := NeedFetch(&m, &begin, &now)
			if err != nil {
				log.Errorln(err)
			}
			log.Info("fetch file ids: %v", fileIDs)
		}()
	}
}

// IsMaterialExist _
func IsMaterialExist(materialID string) bool {
	_, err := ftpclient.GetList("./" + materialID)
	if err != nil {
		if fe, ok := err.(*ftpclient.FTPError); ok {
			fe.Logger()
			return false
		}

		log.Error("[IsMaterialExist] %v", err)
		return false
	}

	return true
}

// FetchMaterialDatas 获取指定文件中的数据
func fetchMaterialDatas(material orm.Material, files []FetchFile) ([]int, error) {
	var fileIDs []int

	if len(files) == 0 {
		return nil, errors.New("没有需要获取的数据文件")
	}

	xr := ftpclient.NewXLSXReader()
	err := xr.ReadSize(resolvePath(material.Name, files[0].File))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("读取数据文件%s失败", files[0].File))
	}
	handleSizePoint(xr.DimSL, material.ID)
	colLen := len(xr.DimSL)

	fileWaitChan := make(chan int, 1)
	xlsxReaders := make([]*ftpclient.XLSXReader, 0)
	for _, f := range files {
		xr := ftpclient.NewXLSXReader()
		path := resolvePath(material.Name, f.File)

		file := orm.GetFileListWithPath(path)
		if file == nil {
			file = &orm.File{Path: path, MaterialID: material.ID, FileDate: f.Date}
			orm.DB.Create(file)
		}
		fileIDs = append(fileIDs, file.ID)
		go func() {
			err := xr.Read(path)
			if err != nil {
				log.Error("read path(%s) error: %v", path, err)
				return
			}
			rowLen := len(xr.DateSet)
			total := (colLen + 1) * rowLen
			orm.DB.Model(&orm.File{}).Where("id = ?", file.ID).Update("total_rows", total)
			xr.PathID = file.ID

			xlsxReaders = append(xlsxReaders, xr)
			fileWaitChan <- 1
		}()
	}

	// 这样做的目的是为了保证拉取数据前，file的total_rows已经准备就绪，否则造成前端数据完成进度条回滚的现象
	var count int
	for {
		i := <-fileWaitChan
		count = count + i
		if count == 3 {
			break
		}
	}

	for _, xr := range xlsxReaders {
		ftpclient.PushStore(xr)
	}

	return fileIDs, nil
}

func resolvePath(m, path string) string {
	return fmt.Sprintf("./%s/%s", m, filepath.Base(path))
}

func handleSizePoint(dimSet map[string]ftpclient.SL, materialID int) {
	tx := orm.DB.Begin()
	for k, v := range dimSet {
		sizeName, pointName := parseSizePoint(k)

		size := orm.GetSizeWithMaterialIDSizeName(sizeName, materialID, tx)
		if size == nil {
			size = &orm.Size{
				Name:       sizeName,
				MaterialID: materialID,
			}
			tx.Create(size)
		}

		point := orm.GetPointWithSizeIDPointName(pointName, size.ID, tx)
		if point == nil {
			point = &orm.Point{
				Name:       pointName,
				SizeID:     size.ID,
				Index:      v.Index,
				UpperLimit: v.USL,
				LowerLimit: v.LSL,
				Norminal:   v.Norminal,
			}
			tx.Create(point)
		} else {
			point.SizeID = size.ID
			point.Index = v.Index
			point.LowerLimit = v.LSL
			point.UpperLimit = v.USL
			point.Norminal = v.Norminal
			tx.Save(point)
		}
	}
	tx.Commit()
}

func parseSizePoint(s string) (string, string) {
	r := strings.Split(s, "_")
	return r[0], s
}

// FetchFile 需要获取的文件
type FetchFile struct {
	File string
	Date time.Time
}

// NeedFetch 判断是否需要从FTP拉取数据
// 给定料号，时间范围，对比数据库中已拉取文件路径，得出是否有需要拉取的文件路径
func NeedFetch(m *orm.Material, begin, end *time.Time) ([]int, error) {
	var files []FetchFile
	var fileIDs []int
	if begin != nil && end != nil {
		if begin.After(*end) {
			return fileIDs, errors.New("时间范围不正确，开始时间不能晚于结束时间")
		}
	}

	ftpFileList, err := ftpclient.GetList("./" + m.Name)
	if err != nil {
		return fileIDs, err
	}

	for _, p := range ftpFileList {
		need, deviceName, fileDate := checkFile(p,begin, end)
		if !need {
			continue
		}

		// 根据文件名创建设备
		createDeviceIfNotExist(deviceName, *m)

		files = append(files, FetchFile{
			File: p,
			Date: *fileDate,
		})
	}

	if len(files) == 0 {
		return fileIDs, nil
	}
	fileIDs, err = fetchMaterialDatas(*m, files)
	if err != nil {
		return fileIDs, err
	}

	return fileIDs, nil
}

func checkFile(fileName string, begin, end *time.Time) (bool, string, *time.Time) {
	var file orm.File
	err := orm.DB.Model(&file).Where("files.path = ?", fileName).First(&file).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Errorln(err)
			return false, "", nil
		}
	}

	if file.ID != 0 {
		return false, "", nil
	}

	reg := regexp.MustCompile(fileNamePattern)
	matched := reg.FindStringSubmatch(fileName)
	if len(matched) <= 4 {
		return false, "", nil
	}
	dateStr := matched[3]
	t, _ := time.Parse(time.RFC3339, fmt.Sprintf("%s-%s-%sT00:00:00+08:00", dateStr[:4], dateStr[4:6], dateStr[6:]))

	if begin == nil || end == nil {
		return true, matched[2], &t
	}
	return begin.Before(t) && end.After(t), matched[2], &t
}

func createDeviceIfNotExist(name string, material orm.Material) {
	device := orm.GetDeviceWithName(name)
	if device == nil || device.MaterialID != material.ID {
		device = &orm.Device{
			Name:       name,
			MaterialID: material.ID,
		}
		orm.DB.Create(device)
	}
}
