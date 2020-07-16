package router

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/log"
	"github.com/SasukeBo/pmes-data-center/router/handler"
	"github.com/gin-gonic/gin"
)

func Start() {
	r := gin.Default()
	//r.Use(cors.Default())

	// Panic Recovery
	r.Use(gin.Recovery())

	// Auth
	auth := r.Group("/auth", handler.HttpRequestLogger())
	{
		auth.POST("/login", handler.Login())
		auth.GET("/logout", handler.Logout())
	}

	// API v1
	api1 := r.Group("/api", handler.HttpRequestLogger(), handler.InjectGinContext())
	{
		api1.POST("/v1", handler.API1())
		api1.POST("/v1/admin", handler.Authenticate(), handler.API1Admin())
	}

	// Active
	//r.GET("/active", handler.Active())

	// Downloads
	r.GET("/downloads/xlsx", handler.DownloadXlsxFile()) // 下载xlsx文件

	// Uploads
	r.MaxMultipartMemory = 256 << 20
	r.POST("/posts", handler.Authenticate(), handler.Post()) // 上传文件

	// Data transfer
	r.POST("/produce", handler.HttpRequestLogger(), handler.DeviceProduce()) // 设备上传生产数据

	log.Info("start service on [%s] mode", configer.GetEnv("env"))
	r.Run(fmt.Sprintf(":%s", configer.GetString("port")))
}
