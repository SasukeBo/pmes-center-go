package main

import (
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/router"
)

func main() {
	if configer.GetString("env") != "dev" {
		go logic.AutoFetch()
		go logic.AutoCleanCacheFile()
	}
	go orm.AutoCacheProducts()
	router.Start()
}
