package main

import (
	"github.com/SasukeBo/ftpviewer/api/v1/admin/logic"
	"github.com/SasukeBo/ftpviewer/router"
)

func main() {
	go logic.AutoFetch()
	router.Start()
}
