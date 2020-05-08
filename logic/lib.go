package logic

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/SasukeBo/ftpviewer/ftpclient"
)

var fileNamePattern = `(.*)-(.*)-(.*)-([w|b])\.xlsx`

// IsMaterialExist _
func IsMaterialExist(materialID string) bool {
	_, err := ftpclient.GetList("./" + materialID)
	if err != nil {
		if fe, ok := err.(*ftpclient.FTPError); ok {
			fe.Logger()
			return false
		}

		log.Printf("[IsMaterialExist] %v", err)
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
				log.Println(err)
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
	var conds []string
	var vars []interface{}
	var files []FetchFile
	var fileIDs []int
	conds = append(conds, "finished = 1")
	conds = append(conds, "material_id = ?")
	vars = append(vars, m.ID)
	if begin != nil && end != nil {
		if begin.After(*end) {
			return fileIDs, errors.New("时间范围不正确，开始时间不能晚于结束时间")
		}
	}
	if begin != nil {
		conds = append(conds, "file_date > ?")
		vars = append(vars, *begin)
	}
	if end != nil {
		conds = append(conds, "file_date < ?")
		vars = append(vars, *end)
	}
	var fetchedFileList []orm.File
	if err := orm.DB.Model(&orm.File{}).Where(strings.Join(conds, " AND "), vars...).Find(&fetchedFileList).Error; err != nil {
		return fileIDs, err
	}

	ftpFileList, err := ftpclient.GetList("./" + m.Name)
	if err != nil {
		return fileIDs, err
	}

	reg := regexp.MustCompilePOSIX(fileNamePattern)
	for _, p := range ftpFileList {
		matched := reg.FindAllStringSubmatch(p, -1)
		if len(matched) == 0 || len(matched[0]) <= 4 {
			continue
		}
		dateStr := matched[0][3]
		fileDate, _ := time.Parse(time.RFC3339, fmt.Sprintf("%s-%s-%sT00:00:00+08:00", dateStr[:4], dateStr[4:6], dateStr[6:]))

		if !fileIsNeed(&fileDate, begin, end) {
			continue
		}

		// 根据文件名创建设备
		createDeviceIfNotExist(matched[0][2], *m)

		fetched := false
		for _, f := range fetchedFileList {
			if strings.Contains(f.Path, p) {
				fetched = true
				break
			}
		}

		if !fetched {
			files = append(files, FetchFile{
				File: p,
				Date: fileDate,
			})
		}
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

func fileIsNeed(fileDate, begin, end *time.Time) bool {
	if begin == nil || end == nil {
		return true
	}
	return begin.Before(*fileDate) && end.After(*fileDate)
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
