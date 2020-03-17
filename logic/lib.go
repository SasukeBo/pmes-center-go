package logic

import (
	"errors"
	"fmt"
	"github.com/SasukeBo/ftpviewer/orm"
	"log"
	"strconv"
	"time"

	"github.com/SasukeBo/ftpviewer/ftpclient"
)

// IsMaterialExist _
func IsMaterialExist(materialID string) bool {
	dirs, err := ftpclient.GetList("./")
	if err != nil {
		if fe, ok := err.(*ftpclient.FTPError); ok {
			fe.Logger()
			return false
		}

		log.Printf("[IsMaterialExist] %v", err)
		return false
	}

	for _, dir := range dirs {
		if dir == materialID {
			return true
		}
	}

	return false
}

// FetchMaterialDatas 根据料号从FTP服务器获取时间范围内数据
func FetchMaterialDatas(materialID string, begin, end *time.Time) (interface{}, error) {
	deviceName := make(map[string]bool)
	fetchList := make([]string)

	cacheDays := 30
	config := orm.GetSystemConfigCache("cache_days")
	if config != nil {
		if v, err := strconv.Atoi(config.Value); err == nil {
			cacheDays = v
		}
	}

	today := time.Now()
	cacheBegin := today.Add(time.Duration(int(time.Hour) * -24 * cacheDays))

	if begin != nil && end != nil {
		if  begin.After(*end) {
			return nil, errors.New("获取数据的开始时间不能在结束时间之后")
		}
	} else if begin == nil && end == nil {
		begin = &cacheBegin
		end = &today
	} else if begin != nil && end == nil {
		if begin.After(today) {
			return nil, nil
		}
		end = &today
	} else if begin == nil && end != nil {
		if end.Before(cacheBegin) {
			return nil, nil
		}
		begin = &cacheBegin
	}

	fileList, err := ftpclient.GetList("./" + materialID)
	if err != nil {
		return nil, err
	}

	for _, filename := range fileList {

	}

	return nil, nil
}
