package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string
	Port string
}

// Server represents the HTTP server
type Server struct {
	config ServerConfig
	router *gin.Engine
}

// NewServer creates a new server instance
func NewServer(config ServerConfig) *Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	return &Server{
		config: config,
		router: router,
	}
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API versioning
	v0 := s.router.Group("/api/v0")
	{
		// VM endpoints
		vms := v0.Group("/vms")
		{
			vms.POST("", CreateVMHandler)
			vms.GET("", ListVMsHandler)
			vms.GET("/:name", GetVMHandler)
			vms.DELETE("/:id", DeleteVMHandler)
		}

		// Node endpoints
		nodes := v0.Group("/nodes")
		{
			nodes.GET("", ListNodesHandler)
			nodes.POST("", AddNodeHandler)
			nodes.GET("/:name", GetNodeHandler)
			nodes.DELETE("/:name", DeleteNodeHandler)
		}
	}

	// Version endpoint
	s.router.GET("/version", VersionHandler)
}

// Start initializes and starts the server
func (s *Server) Start() error {
	s.setupRoutes()

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	log.Printf("Starting server on %s", addr)

	if err := s.router.Run(addr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Serve starts the server with the given configuration
func Serve(host, port string) {
	config := ServerConfig{
		Host: host,
		Port: port,
	}

	server := NewServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// respondWithError sends an error response
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, APIResponse{
		Success: false,
		Error:   message,
	})
}

// respondWithSuccess sends a success response
func respondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// MiddlewareFunc represents a Gin middleware function
type MiddlewareFunc func(*gin.Context)

// withMiddleware adds middleware to a route group
func withMiddleware(group *gin.RouterGroup, middleware ...MiddlewareFunc) {
	for _, m := range middleware {
		group.Use(m)
	}
}

// LoggingMiddleware logs request details
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request details
		log.Printf(
			"%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		)
	}
}

// ErrorMiddleware handles panics and errors
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				respondWithError(c, http.StatusInternalServerError, "Internal server error")
			}
		}()
		c.Next()
	}
}
