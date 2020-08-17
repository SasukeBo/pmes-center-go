package main

import (
	"github.com/SasukeBo/pmes-data-center/api/v1/admin/logic"
	"github.com/SasukeBo/pmes-data-center/router"
)

func main() {
	go logic.AutoFetch()

	router.Start()
}
