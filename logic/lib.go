package logic

import (
	"log"
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
	return nil, nil
}
