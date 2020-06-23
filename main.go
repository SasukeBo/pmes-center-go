package main

import (
	"github.com/SasukeBo/ftpviewer/router"
	"github.com/SasukeBo/ftpviewer/worker"
)

func main() {
	worker.Start()
	router.Start()
}
