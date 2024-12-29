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

	return s.router.Run(addr)
}

// handleIndex handles the main page
func (s *WebServer) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"Title":       "GovnoCloud Dashboard",
		"Description": "Manage your cloud resources",
		"Version":     "v2.0.0",
	})
}
