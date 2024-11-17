package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Listen starts the web server.
func Listen(host, port string) error {
	addr := host + ":" + port
	r := gin.New()
	r.LoadHTMLGlob("templates/*")
	r.Use(cors.Default())
	r.GET("/", RootHandler)
	r.Run(addr)
	return nil
}
