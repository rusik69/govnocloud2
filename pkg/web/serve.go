package web

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Listen(host, port, webPath string) error {
	router := gin.New()
	router.Use(gin.Recovery())

	// Log requests
	router.Use(func(c *gin.Context) {
		log.Printf("%s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	// Redirect root to nodes page
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "nodes")
	})

	router.GET("/vms", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/vms.html", gin.H{})
	})

	router.GET("/containers", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/containers.html", gin.H{})
	})

	router.GET("/dbs", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/dbs.html", gin.H{})
	})

	router.GET("/volumes", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/volumes.html", gin.H{})
	})

	router.GET("/namespaces", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/namespaces.html", gin.H{})
	})

	router.GET("/nodes", func(c *gin.Context) {
		c.HTML(http.StatusOK, webPath+"/nodes.html", gin.H{})
	})
	log.Printf("Starting web server on %s:%s (path: %s)", host, port, webPath)
	return router.Run(host + ":" + port)
}
