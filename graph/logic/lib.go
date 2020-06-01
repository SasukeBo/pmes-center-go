package logic

import (
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"time"
)

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
