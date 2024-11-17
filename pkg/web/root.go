package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RootHandler handles the root route.
func RootHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Govnocloud2",
	})
}
