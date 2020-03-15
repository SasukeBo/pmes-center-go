package main

import (
	"context"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/SasukeBo/ftpviewer/graph"
	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/gin-gonic/gin"
)

func graphqlHandler() gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// GinContextToContextMiddleware store gin.Context into context.Context
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "GinContext", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func main() {
	r := gin.Default()
	r.Use(GinContextToContextMiddleware())
	r.POST("/query", graphqlHandler())
	r.GET("/", gin.BasicAuth(gin.Accounts{
		"sasuke": "Wb922149@...S",
	}), playgroundHandler())
	r.Run(":44761")
}
