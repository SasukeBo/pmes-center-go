package main

import (
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/SasukeBo/ftpviewer/router"
)

func main() {
	go ftpclient.FTPWorker()
	go logic.ClearUp()
	go logic.AutoFetch()
	router.Start()
}
