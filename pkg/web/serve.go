package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"path"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templates embed.FS

var pages = map[string]string{
	"nodes":      "Nodes",
	"vms":        "Virtual Machines",
	"volumes":    "Volumes",
	"containers": "Containers",
	"dbs":        "Databases",
	"namespaces": "Namespaces",
}

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

	// Load templates
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"webpath": func() string { return webPath },
	}).ParseFS(templates, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	// Setup routes
	group := router.Group(webPath)

	// Redirect root to nodes page
	group.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, path.Join(webPath, "nodes"))
	})

	// Setup page routes
	for route, title := range pages {
		route := route
		title := title
		group.GET("/"+route, func(c *gin.Context) {
			data := NewPageData()
			data.Title = "GovnoCloud Dashboard - " + title
			data.Description = "Manage your " + title
			data.WebPath = webPath
			c.HTML(http.StatusOK, route+".html", data)
		})
	}

	log.Printf("Starting web server on %s:%s (path: %s)", host, port, webPath)
	return router.Run(host + ":" + port)
}
