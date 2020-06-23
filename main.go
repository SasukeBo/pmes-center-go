package main

import (
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/router"
	"github.com/SasukeBo/ftpviewer/worker"
)

func main() {
	go ftpclient.FTPWorker()
	// go logic.ClearUp() 不再清除数据，所有数据保持在服务器
	go worker.AutoFetch()
	router.Start()
}
