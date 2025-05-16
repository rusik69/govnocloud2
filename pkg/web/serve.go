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

	// Load templates
	router.LoadHTMLGlob(webPath + "/*.html")

	// Log requests
	router.Use(func(c *gin.Context) {
		log.Printf("%s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Redirect root to nodes page
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":  "GovnoCloud Dashboard",
			"Active": "home",
		})
	})

	router.GET("/nodes", func(c *gin.Context) {
		c.HTML(http.StatusOK, "nodes.html", gin.H{
			"Title":  "Nodes - GovnoCloud",
			"Active": "nodes",
		})
	})

	router.GET("/vms", func(c *gin.Context) {
		c.HTML(http.StatusOK, "vms.html", gin.H{
			"Title":  "VMs - GovnoCloud",
			"Active": "vms",
		})
	})

	router.GET("/containers", func(c *gin.Context) {
		c.HTML(http.StatusOK, "containers.html", gin.H{
			"Title":  "Containers - GovnoCloud",
			"Active": "containers",
		})
	})

	router.GET("/dbs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dbs.html", gin.H{
			"Title":  "Databases - GovnoCloud",
			"Active": "dbs",
		})
	})

	router.GET("/volumes", func(c *gin.Context) {
		c.HTML(http.StatusOK, "volumes.html", gin.H{
			"Title":  "Volumes - GovnoCloud",
			"Active": "volumes",
		})
	})

	router.GET("/namespaces", func(c *gin.Context) {
		c.HTML(http.StatusOK, "namespaces.html", gin.H{
			"Title":  "Namespaces - GovnoCloud",
			"Active": "namespaces",
		})
	})

	log.Printf("Starting web server on %s:%s (path: %s)", host, port, webPath)
	return router.Run(host + ":" + port)
}
