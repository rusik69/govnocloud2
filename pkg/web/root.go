package web

import "github.com/gin-gonic/gin"

// RootHandler handles the root route.
func RootHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "govnocloud2 web",
	})
}
