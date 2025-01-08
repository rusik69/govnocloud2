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

type WebServer struct {
	router  *gin.Engine
	webPath string
}

var pages = map[string]string{
	"nodes":      "Nodes",
	"vms":        "Virtual Machines",
	"volumes":    "Volumes",
	"containers": "Containers",
	"dbs":        "Databases",
	"namespaces": "Namespaces",
}

func NewWebServer(webPath string) *WebServer {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logRequests())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}))

	// Load templates with web path prefix
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"webpath": func() string { return webPath },
	}).ParseFS(templates, "templates/*"))

	router.SetHTMLTemplate(tmpl)

	return &WebServer{
		router:  router,
		webPath: webPath,
	}
}

func (s *WebServer) Start(addr string) error {
	s.setupRoutes()
	log.Printf("Starting web server on %s (path: %s)", addr, s.webPath)
	return s.router.Run(addr)
}

func (s *WebServer) setupRoutes() {
	group := s.router.Group(s.webPath)

	group.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, path.Join(s.webPath, "nodes"))
	})

	for route, title := range pages {
		group.GET("/"+route, s.handlePage(route, title))
	}
}

func (s *WebServer) handlePage(page, title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := NewPageData()
		data.Title = "GovnoCloud Dashboard - " + title
		data.Description = "Manage your " + title
		data.WebPath = s.webPath

		c.HTML(http.StatusOK, page+".html", data)
	}
}

func logRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Web: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

func Listen(host, port, webPath string) error {
	server := NewWebServer(webPath)
	return server.Start(host + ":" + port)
}
