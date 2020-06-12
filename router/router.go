package router

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/router/handler"
	"github.com/SasukeBo/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func Start() {
	r := gin.Default()
	//r.Use(cors.Default())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://192.168.13.104", "http://localhost"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Panic Recovery
	r.Use(gin.Recovery())

	// Auth
	auth := r.Group("/auth")
	{
		auth.POST("/login", handler.Login())
		auth.GET("/logout", handler.Logout())
	}

	// API v1
	api1 := r.Group("/api", handler.Authenticate(), handler.GraphqlResponseLogger(), handler.InjectGinContext())
	{
		api1.POST("/v1", handler.API1())
		api1.POST("/v1/admin", handler.API1Admin())
	}

	// Active
	//r.GET("/active", handler.Active())

	// Downloads
	r.GET("/downloads/xlsx", handler.DownloadXlsxFile()) // 下载xlsx文件

	log.Info("start service on [%s] mode", configer.GetEnv("env"))
	r.Run(fmt.Sprintf(":%s", configer.GetString("port")))
}
