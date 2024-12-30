package web

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templates embed.FS

// WebServer represents the web interface server
type WebServer struct {
	router *gin.Engine
}

// NewWebServer creates a new web interface server
func NewWebServer() *WebServer {
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(config))

	// Load templates
	tmpl := template.Must(template.ParseFS(templates, "templates/*"))
	router.SetHTMLTemplate(tmpl)

	return &WebServer{
		router: router,
	}
}

// Start starts the web interface server
func (s *WebServer) Start(addr string) error {
	// Setup routes
	s.router.GET("/", s.handleIndex)
	s.router.GET("/nodes", s.handleNodes)
	s.router.GET("/vms", s.handleVms)
	s.router.GET("/volumes", s.handleVolumes)
	s.router.GET("/containers", s.handleContainers)
	s.router.GET("/dbs", s.handleDbs)

	return s.router.Run(addr)
}

// handleIndex handles the main page
func (s *WebServer) handleIndex(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/nodes")
}

// handleNodes handles the nodes page
func (s *WebServer) handleNodes(c *gin.Context) {
	c.HTML(http.StatusOK, "nodes.html", gin.H{
		"Title":       "GovnoCloud Dashboard - Nodes",
		"Description": "Manage your cloud nodes",
		"Version":     "v2.0.0",
	})
}

// handleVms handles the VMs page
func (s *WebServer) handleVms(c *gin.Context) {
	c.HTML(http.StatusOK, "vms.html", gin.H{
		"Title":       "GovnoCloud Dashboard - Virtual Machines",
		"Description": "Manage your virtual machines",
		"Version":     "v2.0.0",
	})
}

// handleVolumes handles the volumes page
func (s *WebServer) handleVolumes(c *gin.Context) {
	c.HTML(http.StatusOK, "volumes.html", gin.H{
		"Title":       "GovnoCloud Dashboard - Volumes",
		"Description": "Manage your storage volumes",
		"Version":     "v2.0.0",
	})
}

// handleContainers handles the containers page
func (s *WebServer) handleContainers(c *gin.Context) {
	c.HTML(http.StatusOK, "containers.html", gin.H{
		"Title":       "GovnoCloud Dashboard - Containers",
		"Description": "Manage your containers",
		"Version":     "v2.0.0",
	})
}

// handleDbs handles the databases page
func (s *WebServer) handleDbs(c *gin.Context) {
	c.HTML(http.StatusOK, "dbs.html", gin.H{
		"Title":       "GovnoCloud Dashboard - Databases",
		"Description": "Manage your databases",
		"Version":     "v2.0.0",
	})
}

// Listen starts the web server
func Listen(host, port, path string) error {
	server := NewWebServer()
	return server.Start(host + ":" + port)
}
