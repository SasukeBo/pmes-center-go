package handler

import (
	"bytes"
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	v1generatedadmin "github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	v1resolveradmin "github.com/SasukeBo/pmes-data-center/api/v1/admin/resolver"
	v1generated "github.com/SasukeBo/pmes-data-center/api/v1/generated"
	v1resolver "github.com/SasukeBo/pmes-data-center/api/v1/resolver"
	"github.com/gin-gonic/gin"
)

func API1() gin.HandlerFunc {
	h := handler.NewDefaultServer(v1generated.NewExecutableSchema(v1generated.Config{Resolvers: &v1resolver.Resolver{}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func API1Admin() gin.HandlerFunc {
	h := handler.NewDefaultServer(v1generatedadmin.NewExecutableSchema(v1generatedadmin.Config{Resolvers: &v1resolveradmin.Resolver{}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// injectGinContext inject gin.Context into context.Context
func InjectGinContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "GinContext", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

