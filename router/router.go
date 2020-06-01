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
	r.POST("/api/login", handler.Login())
	r.GET("/api/logout", handler.Logout())

	// API v1
	r.POST("/api/v1", handler.Authenticate(), handler.GraphqlResponseLogger(), handler.InjectGinContext(), handler.GraphqlHandler())

	// Active
	//r.GET("/active", handler.Active())

	// GraphiQL
	r.GET("/", handler.BasicAuth(), handler.PlaygroundHandler())

	// Downloads
	r.GET("/api/downloads", handler.Download())

	log.Info("start service on [%s] mode", configer.GetEnv("env"))
	r.Run(fmt.Sprintf(":%s", configer.GetString("port")))
}
