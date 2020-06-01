package router

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

type object map[string]string

func Start() {
	r := gin.Default()
	//r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://192.168.13.104", "http://localhost"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(gin.Recovery())
	r.POST("/api", authenticate(), graphqlResponseLogger(), injectGinContext(), graphqlHandler())
	r.GET("/active", active)
	r.GET("/", basicAuth, playgroundHandler())
	r.GET("/downloads", download)
	r.Run(fmt.Sprintf(":%s", configer.GetString("port")))
}
