package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// Server represents the HTTP server
type Server struct {
	config types.ServerConfig
	router *gin.Engine
}

var server *Server

var vmManager *VMManager
var containerManager *ContainerManager
var volumeManager *VolumeManager
var namespaceManager *NamespaceManager
var nodeManager *NodeManager
var dbManager *DBManager

// NewServer creates a new server instance
func NewServer(config types.ServerConfig) *Server {
	router := gin.New()
	router.Use(gin.Recovery())
	vmManager = NewVMManager()
	containerManager = NewContainerManager()
	volumeManager = NewVolumeManager()
	namespaceManager = NewNamespaceManager()
	nodeManager = NewNodeManager()
	dbManager = NewDBManager()
	llmManager = NewLLMManager()

	// Configure CORS to allow all origins
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	router.Use(cors.New(corsConfig))

	router.Use(LoggingMiddleware())

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
			vms.POST("/:namespace/:name", CreateVMHandler)
			vms.GET("/:namespace", ListVMsHandler)
			vms.GET("/:namespace/:name", GetVMHandler)
			vms.DELETE("/:namespace/:name", DeleteVMHandler)
			vms.GET("/:namespace/:name/start", StartVMHandler)
			vms.GET("/:namespace/:name/stop", StopVMHandler)
			vms.GET("/:namespace/:name/restart", RestartVMHandler)
			vms.GET("/:namespace/:name/wait", WaitVMHandler)
		}

		// Node endpoints
		nodes := v0.Group("/nodes")
		{
			nodes.GET("/", ListNodesHandler)
			nodes.POST("/", AddNodeHandler)
			nodes.GET("/:name", GetNodeHandler)
			nodes.DELETE("/:name", DeleteNodeHandler)
			nodes.GET("/:name/restart", RestartNodeHandler)
			nodes.GET("/:name/suspend", SuspendNodeHandler)
			nodes.GET("/:name/resume", ResumeNodeHandler)
			nodes.GET("/:name/upgrade", UpgradeNodeHandler)
		}
		dbs := v0.Group("/dbs")
		{
			dbs.GET("/:namespace", ListDBsHandler)
			dbs.POST("/:namespace/:name", CreateDBHandler)
			dbs.GET("/:namespace/:name", GetDBHandler)
			dbs.DELETE("/:namespace/:name", DeleteDBHandler)
		}
		containers := v0.Group("/containers")
		{
			containers.GET("/:namespace", ListContainersHandler)
			containers.POST("/:namespace/:name", CreateContainerHandler)
			containers.GET("/:namespace/:name", GetContainerHandler)
			containers.DELETE("/:namespace/:name", DeleteContainerHandler)
		}
		volumes := v0.Group("/volumes")
		{
			volumes.GET("/:namespace", ListVolumesHandler)
			volumes.POST("/:namespace/:name", CreateVolumeHandler)
			volumes.GET("/:namespace/:name", GetVolumeHandler)
			volumes.DELETE("/:namespace/:name", DeleteVolumeHandler)
		}
		namespaces := v0.Group("/namespaces")
		{
			namespaces.GET("", ListNamespacesHandler)
			namespaces.POST("/:name", CreateNamespaceHandler)
			namespaces.GET("/:name", GetNamespaceHandler)
			namespaces.DELETE("/:name", DeleteNamespaceHandler)
		}
		llms := v0.Group("/llms")
		{
			llms.POST("/:namespace/:name", CreateLLMHandler)
			llms.GET("/:namespace/:name", GetLLMHandler)
			llms.DELETE("/:namespace/:name", DeleteLLMHandler)
			llms.GET("/:namespace", ListLLMsHandler)
		}
		portforward := v0.Group("/portforward")
		{
			portforward.GET("/:type/:namespace/:name/start", PortForwardStartHandler)
			portforward.GET("/:type/:namespace/:name/stop", PortForwardStopHandler)
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
func Serve(serverConfig types.ServerConfig) {
	server = NewServer(serverConfig)
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

// MiddlewareFunc is an alias for gin.HandlerFunc for better readability
type MiddlewareFunc = gin.HandlerFunc

// LoggingMiddleware creates a middleware for request logging
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

// ErrorMiddleware creates a middleware for error handling
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
