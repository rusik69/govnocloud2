package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewWebServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		expected *WebServerConfig
	}{
		{
			name: "default config",
			host: "localhost",
			port: "8080",
			expected: &WebServerConfig{
				Host: "localhost",
				Port: "8080",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewWebServerConfig(tt.host, tt.port)
			if config.Host != tt.expected.Host {
				t.Errorf("expected host %s, got %s", tt.expected.Host, config.Host)
			}
			if config.Port != tt.expected.Port {
				t.Errorf("expected port %s, got %s", tt.expected.Port, config.Port)
			}
		})
	}
}

func TestNewWebServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	config := NewWebServerConfig("localhost", "8080")
	server := NewWebServer(config)

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
	config := NewWebServerConfig("localhost", "8080")
	server := NewWebServer(config)
	server.setupRoutes()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "health check",
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			server.router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestListenWithTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	serverChan := make(chan error, 1)
	
	// Start server in goroutine
	go func() {
		serverChan <- Listen("localhost", "8081")
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test server response
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
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