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
func FetchMaterialDatas(materialID string, begin, end *time.Time) error {
	//deviceName := make(map[string]bool)
	//fetchList := make([]string)

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
			return errors.New("获取数据的开始时间不能在结束时间之后")
		}
	} else if begin == nil && end == nil {
		begin = &cacheBegin
		end = &today
	} else if begin != nil && end == nil {
		if begin.After(today) {
			return errors.New("获取数据时间范围不正确")
		}
		end = &today
	} else if begin == nil && end != nil {
		if end.Before(cacheBegin) {
			return errors.New("获取数据时间范围不正确")
		}
		begin = &cacheBegin
	}

	fileList, err := ftpclient.GetList("./" + materialID)
	if fe, ok := err.(*ftpclient.FTPError); ok {
		fe.Logger()
		return errors.New(fe.Message)
	}

	reg := regexp.MustCompile(fileNamePattern)

	preHandleSize := false
	for _, filename := range fileList {
		matched := reg.FindAllStringSubmatch(filename, -1)
		if len(matched) > 0 && len(matched[0]) > 4 {
			xr := ftpclient.NewXLSXReader()
			err := xr.ReadSize("./" + materialID + "/" + filename)
			if err != nil { // 如果读取失败
				continue
			}
			handleSize(xr.DimSL, materialID)
			preHandleSize = true
			break
		}
	}

	if !preHandleSize {
		return fmt.Errorf("无法为料号%s创建尺寸数据，请检查FTP服务器下料号数据文件格式是否正确！", materialID)
	}

	for _, filename := range fileList {
		matched := reg.FindAllStringSubmatch(filename, -1)
		if len(matched) > 0 && len(matched[0]) > 4 {
			path := "./" + materialID + "/" + filename
			if fileIsNeed(path, matched[0][3], begin, end) {
				createDeviceIfNotExist(matched[0][2], materialID)
				xr := ftpclient.NewXLSXReader()
				err := xr.Read(path)
				if err != nil {
					continue
				}
				orm.DB.Create(&orm.FileList{Path: path})
				ftpclient.PushStore(xr)
			}
		}
	}

	return nil
}

func createDeviceIfNotExist(id, materialID string) {
	deviceName := fmt.Sprintf("%s设备%s", materialID, id)
	device := orm.GetDeviceWithName(deviceName)
	if device == nil {
		device = &orm.Device{
			Name:       deviceName,
			MaterialID: materialID,
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

func handleSize(dimSet map[string]ftpclient.SL, materialID string) {
	tx := orm.DB.Begin()
	for k, v := range dimSet {
		size := orm.GetSizeWithMaterialIDSizeName(k, materialID)
		if size == nil {
			size = &orm.Size{
				Name:       k,
				MaterialID: materialID,
				UpperLimit: v.USL,
				LowerLimit: v.LSL,
			}
			tx.Create(size)
		}
	}
	tx.Commit()
}
