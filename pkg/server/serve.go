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
	r.POST("/api/v0/vms", CreateVMHandler)
	r.GET("/api/v0/vms", ListVMsHandler)
	r.GET("/api/v0/vms/:name", GetVMHandler)
	r.DELETE("/api/v0/vms/:id", DeleteVMHandler)
	r.GET("/api/v0/nodes", ListNodesHandler)
	r.POST("/api/v0/nodes", AddNodeHandler)
	r.GET("/api/v0/nodes/:name", GetNodeHandler)
	r.DELETE("/api/v0/nodes/:name", DeleteNodeHandler)
	r.Run(addr + ":" + port)
}
