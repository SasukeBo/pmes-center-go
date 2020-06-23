package worker

import (
	"github.com/SasukeBo/ftpviewer/data_parser"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/log"
	"time"
)

// autoFetch 自动拉取
func autoFetch() {
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
			err := data_parser.FetchMaterialData(&m)
			if err != nil {
				// TODO: add log
				log.Errorln(err)
			}
		}()
	}
}
