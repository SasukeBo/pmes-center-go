package router

import (
	"bytes"
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/graph/generated"
	"github.com/SasukeBo/ftpviewer/graph/resolver"
	"github.com/gin-gonic/gin"
	"gopkg.in/gookit/color.v1"
	"io/ioutil"
)

func graphqlHandler() gin.HandlerFunc {
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{}}))

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

// injectGinContext inject gin.Context into context.Context
func injectGinContext() gin.HandlerFunc {
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

func graphqlResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if configer.GetEnv("env") == "prod" {
			c.Next()
			return
		}

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = rw
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Next()
		fmt.Printf("\n%s\n", color.Warn.Render("[Debug GraphQL]"))
		fmt.Printf("%s %s\n", color.Notice.Render("[Request Body]"), string(body))
		fmt.Printf("%s %s\n\n", color.Notice.Render("[Response Body]"), rw.body.String())
	}
}

var basicAuth = gin.BasicAuth(gin.Accounts{"sasuke": "Wb922149@...S"})
