package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"time"

	//"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/SasukeBo/ftpviewer/ftpclient"
	"github.com/SasukeBo/ftpviewer/graph"
	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/SasukeBo/ftpviewer/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func graphqlHandler() gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/api")
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GinContextToContextMiddleware store gin.Context into context.Context
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//if err := logic.ValidateExpired(); err != nil {
		//	e := err.(*gqlerror.Error)
		//	c.Header("content-type", "application/json")
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"errors": []interface{}{e}})
		//	return
		//}
		ctx := context.WithValue(c.Request.Context(), "GinContext", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func main() {
	go ftpclient.FTPWorker()
	go logic.ClearUp()
	r := gin.Default()
	//r.Use(cors.Default())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://192.168.13.156"},
		AllowMethods:     []string{"POST"},
		AllowHeaders:     []string{"Origin", "content-type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))
	r.Use(gin.Recovery())
	r.POST("/api", GinContextToContextMiddleware(), graphqlHandler())
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
	r.Run(":44761")
}
