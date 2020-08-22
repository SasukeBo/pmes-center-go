package handler

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	v1generatedadmin "github.com/SasukeBo/pmes-data-center/api/v1/admin/generated"
	v1resolveradmin "github.com/SasukeBo/pmes-data-center/api/v1/admin/resolver"
	v1generated "github.com/SasukeBo/pmes-data-center/api/v1/generated"
	v1resolver "github.com/SasukeBo/pmes-data-center/api/v1/resolver"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/gin-gonic/gin"
)

func API1() gin.HandlerFunc {
	h := handler.NewDefaultServer(v1generated.NewExecutableSchema(v1generated.Config{Resolvers: &v1resolver.Resolver{}}))

	return func(c *gin.Context) {
		request, ok := c.Get("requestBody")
		var key string
		if ok {
			key = fmt.Sprintf("%x-request", md5.Sum([]byte(fmt.Sprint(request))))
			if value, err := cache.Get(key); err == nil {
				c.Writer.Write([]byte(value))
				c.Writer.Header().Set("Content-Type", "application/json")
				return
			}
		}
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			Body:           bytes.NewBufferString(""),
		}
		c.Writer = rw
		h.ServeHTTP(c.Writer, c.Request)
		if key != "" {
			cache.Set(key, rw.Body.String())
		}
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
	Body *bytes.Buffer
}

func (rw responseWriter) Write(b []byte) (int, error) {
	rw.Body.Write(b)
	return rw.ResponseWriter.Write(b)
}
