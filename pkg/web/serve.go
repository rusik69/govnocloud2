package web

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Serve starts the web server.
func Serve() error {
	r := gin.New()
	r.Use(cors.Default())
	r.GET("/", RootHandler)
	return nil
}
