package logic

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"regexp"
	"strconv"
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

// FetchMaterialDatas 根据料号从FTP服务器获取时间范围内数据
func FetchMaterialDatas(material orm.Material, begin, end *time.Time) ([]int, error) {
	var fileIDs []int
	cacheDays := 30
	config := orm.GetSystemConfig("cache_days")
	if config != nil {
		if v, err := strconv.Atoi(config.Value); err == nil {
			cacheDays = v
		}
	}

	today := time.Now()
	cacheBegin := today.Add(time.Duration(int(time.Hour) * -24 * cacheDays))

	if begin != nil && end != nil {
		if begin.After(*end) {
			return fileIDs, errors.New("获取数据的开始时间不能在结束时间之后")
		}
	} else if begin == nil && end == nil {
		begin = &cacheBegin
		end = &today
	} else if begin != nil && end == nil {
		if begin.After(today) {
			return fileIDs, errors.New("获取数据时间范围不正确")
		}
		end = &today
	} else if begin == nil && end != nil {
		if end.Before(cacheBegin) {
			return fileIDs, errors.New("获取数据时间范围不正确")
		}
		begin = &cacheBegin
	}

	fileList, err := ftpclient.GetList("./" + material.Name)
	if fe, ok := err.(*ftpclient.FTPError); ok {
		fe.Logger()
		return fileIDs, errors.New(fe.Message)
	}

	reg := regexp.MustCompile(fileNamePattern)

	// preHandleSize := false
	for _, filename := range fileList {
		matched := reg.FindAllStringSubmatch(filename, -1)
		if len(matched) > 0 && len(matched[0]) > 4 {
			xr := ftpclient.NewXLSXReader()
			err := xr.ReadSize("./" + material.Name + "/" + filename)
			if err != nil { // 如果读取失败
				continue
			}
			// preHandleSize = true
			handleSize(xr.DimSL, material.ID)
			break
		}
	}

	for _, filename := range fileList {
		matched := reg.FindAllStringSubmatch(filename, -1)
		if len(matched) > 0 && len(matched[0]) > 4 {
			path := "./" + material.Name + "/" + filename
			if fileIsNeed(path, matched[0][3], begin, end) {
				createDeviceIfNotExist(matched[0][2], material)
				xr := ftpclient.NewXLSXReader()
				fileList := orm.FileList{Path: path, MaterialID: material.ID}
				orm.DB.Create(&fileList)
				fileIDs = append(fileIDs, fileList.ID)
				go func() {
					err := xr.Read(path)
					if err != nil {
						return
					}
					xr.PathID = fileList.ID
					ftpclient.PushStore(xr)
				}()
			}
		}
	}

	return fileIDs, nil
}

func createDeviceIfNotExist(id string, material orm.Material) {
	deviceName := fmt.Sprintf("%s设备%s", material.Name, id)
	device := orm.GetDeviceWithName(deviceName)
	if device == nil {
		device = &orm.Device{
			Name:       deviceName,
			MaterialID: material.ID,
		}
		orm.DB.Create(device)
	}
}

func fileIsNeed(path, name string, begin, end *time.Time) bool {
	if fl := orm.GetFileListWithPath(path); fl != nil {
		return false
	}
	if begin == nil || end == nil {
		return true
	}
	t, _ := time.Parse(time.RFC3339, fmt.Sprintf("%s-%s-%sT00:00:00+08:00", name[:4], name[4:6], name[6:]))
	return begin.Before(t) && end.After(t)
}

func handleSize(dimSet map[string]ftpclient.SL, materialID int) {
	tx := orm.DB.Begin()
	for k, v := range dimSet {
		size := orm.GetSizeWithMaterialIDSizeName(k, materialID)
		if size == nil {
			size = &orm.Size{
				Name:       k,
				Index:      v.Index,
				MaterialID: materialID,
				UpperLimit: v.USL,
				LowerLimit: v.LSL,
			}
			tx.Create(size)
		} else if size.Index != v.Index {
			tx.Model(size).Update("index", v.Index)
		}
	}
	tx.Commit()
}
