package server

import "github.com/gin-gonic/gin"

// VersionHandler returns the server version.
func VersionHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": "v0.0.1",
	})
}
