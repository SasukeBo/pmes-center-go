package main

import (
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/router"
)

func main() {
	go ftpclient.FTPWorker()
	// go logic.ClearUp() 不再清除数据，所有数据保持在服务器
	go logic.AutoFetch()
	router.Start()
}
