package web

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templates embed.FS

type WebServer struct {
	router *gin.Engine
}

var pages = map[string]string{
	"nodes":      "Nodes",
	"vms":        "Virtual Machines",
	"volumes":    "Volumes",
	"containers": "Containers",
	"dbs":        "Databases",
	"namespaces": "Namespaces",
}

func NewWebServer() *WebServer {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logRequests())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	tmpl := template.Must(template.ParseFS(templates, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	return &WebServer{router: router}
}

func (s *WebServer) Start(addr string) error {
	s.setupRoutes()
	log.Printf("Starting web server on %s", addr)
	return s.router.Run(addr)
}

func (s *WebServer) setupRoutes() {
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/nodes")
	})

	for route, title := range pages {
		s.router.GET("/"+route, s.handlePage(route, title))
	}
}

func (s *WebServer) handlePage(page, title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := NewPageData()
		data.Title = "GovnoCloud Dashboard - " + title
		data.Description = "Manage your " + title

		c.HTML(http.StatusOK, page+".html", data)
	}
}

func logRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Web: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

func Listen(host, port string) error {
	server := NewWebServer()
	return server.Start(host + ":" + port)
}
