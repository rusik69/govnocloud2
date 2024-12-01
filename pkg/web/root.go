package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PageData represents the data passed to HTML templates
type PageData struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// NewPageData creates a new page data instance with defaults
func NewPageData() *PageData {
	return &PageData{
		Title: "Govnocloud2",
		Data:  make(map[string]interface{}),
	}
}

// RootHandler handles the root route
func RootHandler(c *gin.Context) {
	data := NewPageData()
	data.Description = "Cloud Management Platform"
	data.Version = GetVersion()

	// Add any additional data needed for the root page
	data.Data["features"] = []string{
		"VM Management",
		"Database Management",
		"Node Management",
	}

	renderTemplate(c, "index.html", data)
}

// renderTemplate renders an HTML template with error handling
func renderTemplate(c *gin.Context, template string, data interface{}) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("X-Content-Type-Options", "nosniff")

	if err := c.HTML(http.StatusOK, template, data); err != nil {
		handleTemplateError(c, err)
	}
}

// handleTemplateError handles template rendering errors
func handleTemplateError(c *gin.Context, err error) {
	// Log the error
	logger := GetLogger()
	logger.Error("Template rendering error", "error", err)

	// Return a generic error page
	c.HTML(http.StatusInternalServerError, "error.html", gin.H{
		"title": "Error",
		"error": "An error occurred while rendering the page",
	})
}

// GetVersion returns the current version
func GetVersion() string {
	// This could be set during build time
	return "v0.0.1"
}

// GetLogger returns the logger instance
func GetLogger() Logger {
	// This could be replaced with a proper logger implementation
	return defaultLogger{}
}

// Logger interface for logging
type Logger interface {
	Error(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

// defaultLogger implements Logger interface
type defaultLogger struct{}

func (l defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	// Implement proper error logging
}

func (l defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	// Implement proper info logging
}

func (l defaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	// Implement proper debug logging
}

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(c *gin.Context) {
	c.HTML(http.StatusNotFound, "error.html", gin.H{
		"title": "Page Not Found",
		"error": "The requested page could not be found",
	})
}

// ErrorHandler handles internal server errors
func ErrorHandler(c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "error.html", gin.H{
		"title": "Internal Server Error",
		"error": "An unexpected error occurred",
	})
}
