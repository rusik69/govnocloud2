package client

import (
	"fmt"
	"net/http"
	"time"
)

// Client represents an HTTP client for API operations
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(host, port, username, password string) *Client {
	return &Client{
		baseURL:  fmt.Sprintf("http://%s:%s/api/v0", host, port),
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}
