package web

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// WebServerConfig holds configuration for the web server
type WebServerConfig struct {
	Host         string
	Port         string
	TemplatePath string
	StaticPath   string
	Debug        bool
}

// WebServer represents the web server instance
type WebServer struct {
	config *WebServerConfig
	router *gin.Engine
	logger Logger
}

// NewWebServerConfig creates a new web server configuration with defaults
func NewWebServerConfig(host, port string) *WebServerConfig {
	return &WebServerConfig{
		Host:         host,
		Port:         port,
		TemplatePath: "templates/*",
		StaticPath:   "static",
		Debug:        false,
	}
}

// NewWebServer creates a new web server instance
func NewWebServer(config *WebServerConfig) *WebServer {
	// Set gin mode based on debug setting
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add CORS middleware
	router.Use(cors.Default())

	return &WebServer{
		config: config,
		router: router,
		logger: GetLogger(),
	}
}

// setupMiddleware configures middleware for the server
func (s *WebServer) setupMiddleware() {
	// Add CORS middleware
	s.router.Use(cors.Default())

	// Add logging middleware
	s.router.Use(LoggingMiddleware())

	// Add error handling middleware
	s.router.Use(ErrorMiddleware())
}

// setupTemplates configures template rendering
func (s *WebServer) setupTemplates() error {
	// Load HTML templates from the configured path
	s.router.LoadHTMLGlob(s.config.TemplatePath)
	return nil
}

// setupStaticFiles configures static file serving
func (s *WebServer) setupStaticFiles() {
	if s.config.StaticPath != "" {
		s.router.Static("/static", s.config.StaticPath)
		s.router.StaticFile("/favicon.ico", filepath.Join(s.config.StaticPath, "favicon.ico"))
	}
}

// setupRoutes configures all routes for the server
func (s *WebServer) setupRoutes() {
	// Root route
	s.router.GET("/", RootHandler)

	// Health check
	s.router.GET("/health", HealthHandler)

	// Error handlers
	s.router.NoRoute(NotFoundHandler)
	s.router.NoMethod(MethodNotAllowedHandler)
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler(c *gin.Context) {
	c.HTML(http.StatusMethodNotAllowed, "error.html", gin.H{
		"title": "Method Not Allowed",
		"error": "The requested method is not allowed for this endpoint",
	})
}

// Start initializes and starts the web server
func (s *WebServer) Start() error {
	// Setup server components
	s.setupMiddleware()

	if err := s.setupTemplates(); err != nil {
		return err
	}

	s.setupStaticFiles()
	s.setupRoutes()

	// Log startup
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	s.logger.Info("Starting web server", "address", addr)

	// Start server
	if err := s.router.Run(addr); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// Listen starts the web server with the given configuration
func Listen(host, port string) error {
	// Set Gin mode based on environment
	if os.Getenv("DEBUG") == "true" {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in debug mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
		log.Println("Running in release mode")
	}

	router := gin.New()

	// Add debug middleware in debug mode
	if gin.Mode() == gin.DebugMode {
		router.Use(gin.Logger())   // Detailed request logging
		router.Use(gin.Recovery()) // Panic recovery
	}

	// CORS configuration
	router.Use(cors.Default())

	// Load HTML templates
	router.LoadHTMLGlob("pkg/web/templates/*")

	// Setup routes
	router.GET("/", RootHandler)
	router.GET("/health", HealthHandler)
	router.NoRoute(NotFoundHandler)

	// Start server
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Web server listening on %s", addr)
	return router.Run(addr)
}

// LoggingMiddleware creates a middleware for request logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log request details
		log.Printf(
			"%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			c.Request.UserAgent(),
		)
	}
}

// ErrorMiddleware creates a middleware for error handling
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.HTML(http.StatusInternalServerError, "error.html", gin.H{
					"title": "Internal Server Error",
					"error": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
