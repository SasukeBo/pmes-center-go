package main

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/log"
	handler2 "github.com/SasukeBo/pmes-data-center/realtime_device/handler"
	"github.com/SasukeBo/pmes-data-center/router/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Panic Recovery
	r.Use(gin.Recovery())

	// Data transfer
	r.POST("/produce", handler.HttpRequestLogger(), handler2.DeviceProduce()) // 设备上传生产数据

	log.Info("start service on [%s] mode", configer.GetEnv("env"))
	r.Run(fmt.Sprintf(":%s", configer.GetString("realtime_device_listen_port")))
}
