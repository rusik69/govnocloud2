package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Serve starts the web server.
func Serve(host, port string) error {
	addr := host + ":" + port
	r := gin.New()
	r.Use(cors.Default())
	r.GET("/", RootHandler)
	r.Run(addr)
	return nil
}
