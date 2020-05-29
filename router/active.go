package router

import (
	"github.com/SasukeBo/ftpviewer/graph/logic"
	"github.com/gin-gonic/gin"
	"net/http"
)

func active(c *gin.Context) {
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
		"message": "active",
	})
}
