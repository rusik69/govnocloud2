package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		expected ServerConfig
	}{
		{
			name: "default config",
			host: "localhost",
			port: "8080",
			expected: ServerConfig{
				Host: "localhost",
				Port: "8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ServerConfig{
				Host: tt.host,
				Port: tt.port,
			}
			if config.Host != tt.expected.Host {
				t.Errorf("expected host %s, got %s", tt.expected.Host, config.Host)
			}
			if config.Port != tt.expected.Port {
				t.Errorf("expected port %s, got %s", tt.expected.Port, config.Port)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := ServerConfig{
		Host: "localhost",
		Port: "8080",
	}
	server := NewServer(config)

	if server.config != config {
		t.Error("server config not set correctly")
	}
	if server.router == nil {
		t.Error("router not initialized")
	}
}

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	LoggingMiddleware()(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	handler := ErrorMiddleware()
	handler(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := ServerConfig{
		Host: "localhost",
		Port: "8080",
	}
	server := NewServer(config)
	server.setupRoutes()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "version endpoint",
			path:           "/version",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "list VMs",
			path:           "/api/v0/vms",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list nodes",
			path:           "/api/v0/nodes",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list databases",
			path:           "/api/v0/dbs",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			server.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestServe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serverChan := make(chan error, 1)
	
	// Start server in goroutine
	go func() {
		Serve("localhost", "8082")
		serverChan <- nil
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test endpoints
	endpoints := []string{
		"/version",
		"/api/v0/vms",
		"/api/v0/nodes",
		"/api/v0/dbs",
	}

	for _, endpoint := range endpoints {
		resp, err := http.Get("http://localhost:8082" + endpoint)
		if err != nil {
			t.Fatalf("failed to connect to server at %s: %v", endpoint, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: expected status 200, got %d", endpoint, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Check for server errors
	select {
	case err := <-serverChan:
		if err != nil {
			t.Errorf("server error: %v", err)
		}
	default:
		// Server is running normally
	}
}
