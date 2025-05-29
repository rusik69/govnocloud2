package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
	"golang.org/x/time/rate"
)

// Server represents the HTTP server
type Server struct {
	config  types.ServerConfig
	router  *gin.Engine
	limiter *rate.Limiter
}

var server *Server

var vmManager *VMManager
var containerManager *ContainerManager
var volumeManager *VolumeManager
var namespaceManager *NamespaceManager
var nodeManager *NodeManager
var postgresManager *PostgresManager
var mysqlManager *MysqlManager
var clickhouseManager *ClickhouseManager
var llmManager *LLMManager
var userManager *UserManager

// NewServer creates a new server instance
func NewServer(config types.ServerConfig) *Server {
	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize rate limiter (100 requests per minute)
	limiter := rate.NewLimiter(rate.Every(time.Second/100), 100)

	// Initialize managers
	vmManager = NewVMManager()
	containerManager = NewContainerManager()
	volumeManager = NewVolumeManager()
	namespaceManager = NewNamespaceManager()
	nodeManager = NewNodeManager()
	postgresManager = NewPostgresManager()
	mysqlManager = NewMysqlManager()
	clickhouseManager = NewClickhouseManager()
	llmManager = NewLLMManager()
	userManager = NewUserManager()

	// Configure CORS with more restrictive settings
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8080", "http://127.0.0.1:8080", "http://master.govno2.cloud:8080"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = true
	corsConfig.MaxAge = 12 * time.Hour
	router.Use(cors.New(corsConfig))

	router.Use(LoggingMiddleware())
	router.Use(ErrorMiddleware())

	return &Server{
		config:  config,
		router:  router,
		limiter: limiter,
	}
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API versioning
	v0 := s.router.Group("/api/v0")
	{
		// Public endpoints (no auth required)
		v0.GET("/version", VersionHandler)

		// Protected endpoints (require authentication)
		protected := v0.Group("")
		{
			// VM endpoints
			vms := protected.Group("/vms")
			{
				vms.POST("/:namespace/:name", s.ValidateNamespaceAccess(), CreateVMHandler)
				vms.GET("/:namespace", s.ValidateNamespaceAccess(), ListVMsHandler)
				vms.GET("/:namespace/:name", s.ValidateNamespaceAccess(), GetVMHandler)
				vms.DELETE("/:namespace/:name", s.ValidateNamespaceAccess(), DeleteVMHandler)
				vms.GET("/:namespace/:name/start", s.ValidateNamespaceAccess(), StartVMHandler)
				vms.GET("/:namespace/:name/stop", s.ValidateNamespaceAccess(), StopVMHandler)
				vms.GET("/:namespace/:name/restart", s.ValidateNamespaceAccess(), RestartVMHandler)
				vms.GET("/:namespace/:name/wait", s.ValidateNamespaceAccess(), WaitVMHandler)
			}

			// Node endpoints
			nodes := protected.Group("/nodes")
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
			postgres := protected.Group("/postgres")
			{
				postgres.GET("/:namespace", ListPostgresHandler)
				postgres.POST("/:namespace/:name", CreatePostgresHandler)
				postgres.GET("/:namespace/:name", GetPostgresHandler)
				postgres.DELETE("/:namespace/:name", DeletePostgresHandler)
			}
			mysql := protected.Group("/mysql")
			{
				mysql.GET("/:namespace", ListMysqlHandler)
				mysql.POST("/:namespace/:name", CreateMysqlHandler)
				mysql.GET("/:namespace/:name", GetMysqlHandler)
				mysql.DELETE("/:namespace/:name", DeleteMysqlHandler)
			}
			clickhouse := protected.Group("/clickhouse")
			{
				clickhouse.GET("/:namespace", ListClickhouseHandler)
				clickhouse.POST("/:namespace/:name", CreateClickhouseHandler)
				clickhouse.GET("/:namespace/:name", GetClickhouseHandler)
				clickhouse.DELETE("/:namespace/:name", DeleteClickhouseHandler)
			}
			containers := protected.Group("/containers")
			{
				containers.GET("/:namespace", ListContainersHandler)
				containers.POST("/:namespace/:name", CreateContainerHandler)
				containers.GET("/:namespace/:name", GetContainerHandler)
				containers.DELETE("/:namespace/:name", DeleteContainerHandler)
			}
			volumes := protected.Group("/volumes")
			{
				volumes.GET("/:namespace", ListVolumesHandler)
				volumes.POST("/:namespace/:name", CreateVolumeHandler)
				volumes.GET("/:namespace/:name", GetVolumeHandler)
				volumes.DELETE("/:namespace/:name", DeleteVolumeHandler)
			}
			llms := protected.Group("/llms")
			{
				llms.POST("/:namespace/:name", CreateLLMHandler)
				llms.GET("/:namespace/:name", GetLLMHandler)
				llms.DELETE("/:namespace/:name", DeleteLLMHandler)
				llms.GET("/:namespace", ListLLMsHandler)
			}
			namespaces := protected.Group("/namespaces")
			{
				namespaces.GET("", ListNamespacesHandler)
				namespaces.POST("/:name", CreateNamespaceHandler)
				namespaces.GET("/:name", GetNamespaceHandler)
				namespaces.DELETE("/:name", DeleteNamespaceHandler)
			}
			users := protected.Group("/users")
			{
				users.GET("", ListUsersHandler)
				users.POST("/:name", CreateUserHandler)
				users.GET("/:name", GetUserHandler)
				users.DELETE("/:name", DeleteUserHandler)
				users.POST("/:name/password", SetUserPasswordHandler)
				users.POST("/:name/namespaces/:namespace", AddNamespaceToUserHandler)
				users.DELETE("/:name/namespaces/:namespace", RemoveNamespaceFromUserHandler)
			}
		}
	}
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
	defer userManager.etcdClient.Close()

	err := CreateRootUser()
	if err != nil {
		log.Fatalf("failed to create root user: %v", err)
	}

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
	log.Printf("responding with error: %s", message)
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

// ValidateNamespaceAccess checks if the user has access to the requested namespace
func (s *Server) ValidateNamespaceAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			respondWithError(c, http.StatusUnauthorized, "User not found in context")
			c.Abort()
			return
		}

		namespace := c.Param("namespace")
		if !userManager.HasNamespaceAccess(user.(*types.User), namespace) {
			respondWithError(c, http.StatusForbidden, "Access to namespace denied")
			c.Abort()
			return
		}
		c.Next()
	}
}
