package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Version information
const (
	Version     = "v0.0.1"
	APIVersion  = "v0"
	BuildCommit = "dev"
	BuildTime   = "unknown"
)

// VersionInfo represents version information
type VersionInfo struct {
	Version     string `json:"version"`
	APIVersion  string `json:"apiVersion"`
	BuildCommit string `json:"buildCommit,omitempty"`
	BuildTime   string `json:"buildTime,omitempty"`
}

// GetVersionInfo returns the current version information
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:     Version,
		APIVersion:  APIVersion,
		BuildCommit: BuildCommit,
		BuildTime:   BuildTime,
	}
}

// VersionHandler returns the server version information
func VersionHandler(c *gin.Context) {
	respondWithSuccess(c, GetVersionInfo())
}

// VersionResponse represents the version endpoint response
type VersionResponse struct {
	APIResponse
	Data VersionInfo `json:"data"`
}

// VersionMiddleware adds version header to responses
func VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", APIVersion)
		c.Header("X-App-Version", Version)
		c.Next()
	}
}

// CheckVersion validates API version compatibility
func CheckVersion(requiredVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requiredVersion != APIVersion {
			respondWithError(c, http.StatusBadRequest, "API version mismatch")
			c.Abort()
			return
		}
		c.Next()
	}
}
