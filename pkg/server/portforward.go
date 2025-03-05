package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PortForwardStartHandler is the handler for starting a port forward
func PortForwardStartHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Port forward started"})
}

// PortForwardStopHandler is the handler for stopping a port forward
func PortForwardStopHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Port forward stopped"})
}
