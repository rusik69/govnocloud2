package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
)

func TestGetServerVersion(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse client.VersionResponse
		statusCode     int
		expectedErr    bool
		expectedVer    string
	}{
		{
			name: "successful version fetch",
			serverResponse: client.VersionResponse{
				Version: "v0.0.1",
			},
			statusCode:  http.StatusOK,
			expectedErr: false,
			expectedVer: "v0.0.1",
		},
		{
			name: "server error",
			serverResponse: client.VersionResponse{
				Version: "",
			},
			statusCode:  http.StatusInternalServerError,
			expectedErr: true,
			expectedVer: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request path
				if r.URL.Path != "/version" {
					t.Errorf("expected path /version, got %s", r.URL.Path)
				}

				// Set response status code
				w.WriteHeader(tt.statusCode)

				// Write response
				if tt.statusCode == http.StatusOK {
					err := json.NewEncoder(w).Encode(tt.serverResponse)
					if err != nil {
						t.Errorf("failed to encode response: %v", err)
					}
				}
			}))
			defer server.Close()

			// Extract host and port from test server
			host := server.URL[7:] // Remove "http://"
			var port string
			for i := len(host) - 1; i >= 0; i-- {
				if host[i] == ':' {
					port = host[i+1:]
					host = host[:i]
					break
				}
			}

			// Call the function
			ver, err := client.GetServerVersion(host, port)

			// Check error
			if (err != nil) != tt.expectedErr {
				t.Errorf("GetServerVersion() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			// Check version
			if !tt.expectedErr && ver != tt.expectedVer {
				t.Errorf("GetServerVersion() = %v, want %v", ver, tt.expectedVer)
			}
		})
	}
}
