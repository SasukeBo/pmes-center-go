package handler

import (
	"bytes"
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/gin-gonic/gin"
	"gopkg.in/gookit/color.v1"
	"io/ioutil"
)

func HttpRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if false && configer.GetEnv("env") == "prod" {
			//if true || configer.GetEnv("env") == "prod" {
			c.Next()
			return
		}

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			Body:           bytes.NewBufferString(""),
		}
		c.Writer = rw
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Set("requestBody", string(body))
		c.Next()
		fmt.Printf("\n%s\n", color.Warn.Render("[Debug Output]"))
		fmt.Printf("%s %s\n", color.Notice.Render("[Request Body]"), string(body))
		fmt.Printf("%s %s\n\n", color.Notice.Render("[Response Body]"), rw.Body.String())
	}
}
