package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Serve starts the server.
func Serve(addr, port string) {
	// Start the server
	r := gin.New()
	r.Use(cors.Default())
	r.GET("/version", VersionHandler)
	r.POST("/api/v1/vms", CreateVMHandler)
	r.Run(addr + ":" + port)
}
