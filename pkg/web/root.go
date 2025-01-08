package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PageData represents the data passed to HTML templates
type PageData struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	MasterHost  string                 `json:"master_host"`
	MasterPort  string                 `json:"master_port"`
}

// NewPageData creates a new page data instance with defaults
func NewPageData() PageData {
	return PageData{
		Title: "Govnocloud2",
		Data:  make(map[string]interface{}),
	}
}

// RootHandler handles the root route
func RootHandler(c *gin.Context) {
	data := NewPageData()
	data.Description = "Cloud Management Platform"
	data.Data["features"] = []string{
		"VM Management",
		"Container Management",
		"Database Management",
		"Node Management",
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

// HandleError renders the error page with the given status and message
func HandleError(c *gin.Context, status int, message string) {
	c.HTML(status, "error.html", gin.H{
		"title": "Error",
		"error": message,
	})
}

// NotFoundHandler handles 404 errors
func NotFoundHandler(c *gin.Context) {
	HandleError(c, http.StatusNotFound, "The requested page could not be found")
}

// ErrorHandler handles internal server errors
func ErrorHandler(c *gin.Context) {
	HandleError(c, http.StatusInternalServerError, "An unexpected error occurred")
}
