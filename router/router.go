package router

import (
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
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
	r.POST("/api", graphqlResponseLogger(), ginContextToContextMiddleware(), graphqlHandler())
	basicAuth := gin.BasicAuth(gin.Accounts{
		"sasuke": "Wb922149@...S",
	})
	r.GET("/active", func(c *gin.Context) {
		token := c.Query("active_token")
		if err := logic.Active(token); err != nil {
			c.Header("content-type", "application/json")
			c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"status":  "failed",
				"message": err.Error(),
			})
			return
		}

		c.Header("content-type", "application/json")
		c.AbortWithStatusJSON(http.StatusOK, map[string]interface{}{
			"status":  "ok",
			"message": "actived",
		})
	})
	r.GET("/", basicAuth, playgroundHandler())
	r.GET("/downloads", download)
	r.Run(":44761")
}
