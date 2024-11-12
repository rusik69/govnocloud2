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
	r.GET("/api/v1/vms", ListVMsHandler)
	r.GET("/api/v1/vms/:name", GetVMHandler)
	r.DELETE("/api/v1/vms/:id", DeleteVMHandler)
	r.Run(addr + ":" + port)
}
