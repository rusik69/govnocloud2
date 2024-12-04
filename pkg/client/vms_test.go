package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rusik69/govnocloud2/pkg/client"
	"github.com/rusik69/govnocloud2/pkg/types"
)

func setupTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *client.Client) {
	server := httptest.NewServer(handler)

	// Extract host and port from test server URL
	host := server.URL[7:] // Remove "http://"
	var port string
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			port = host[i+1:]
			host = host[:i]
			break
		}
	}

	return server, client.NewClient(host, port)
}

func TestCreateVM(t *testing.T) {
	tests := []struct {
		name        string
		vmName      string
		image       string
		size        string
		namespace   string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful creation",
			vmName:      "test-vm",
			image:       "ubuntu",
			size:        "small",
			namespace:   "default",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "server error",
			vmName:      "test-vm",
			image:       "ubuntu",
			size:        "small",
			namespace:   "default",
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/vms" {
					t.Errorf("expected path /api/v1/vms, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
			})
			defer server.Close()

			err := c.CreateVM(tt.vmName, tt.image, tt.size, tt.namespace)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateVM() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestListVMs(t *testing.T) {
	expectedVMs := []types.VM{
		{Name: "vm1", Image: "ubuntu", Size: "small", Namespace: "default"},
		{Name: "vm2", Image: "debian", Size: "medium", Namespace: "test"},
	}

	tests := []struct {
		name        string
		vms         []types.VM
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful list",
			vms:         expectedVMs,
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "server error",
			vms:         nil,
			statusCode:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/api/v1/vms" {
					t.Errorf("expected path /api/v1/vms, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					err := json.NewEncoder(w).Encode(tt.vms)
					if err != nil {
						t.Errorf("failed to encode response: %v", err)
					}
				}
			})
			defer server.Close()

			vms, err := c.ListVMs()
			if (err != nil) != tt.expectError {
				t.Errorf("ListVMs() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && len(vms) != len(tt.vms) {
				t.Errorf("ListVMs() got %v VMs, want %v", len(vms), len(tt.vms))
			}
		})
	}
}

func TestGetVM(t *testing.T) {
	expectedVM := types.VM{
		Name:      "test-vm",
		Image:     "ubuntu",
		Size:      "small",
		Namespace: "default",
	}

	tests := []struct {
		name        string
		vmName      string
		vm          *types.VM
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful get",
			vmName:      "test-vm",
			vm:          &expectedVM,
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "not found",
			vmName:      "nonexistent",
			vm:          nil,
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					err := json.NewEncoder(w).Encode(tt.vm)
					if err != nil {
						t.Errorf("failed to encode response: %v", err)
					}
				}
			})
			defer server.Close()

			vm, err := c.GetVM(tt.vmName)
			if (err != nil) != tt.expectError {
				t.Errorf("GetVM() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && vm.Name != tt.vm.Name {
				t.Errorf("GetVM() got %v, want %v", vm.Name, tt.vm.Name)
			}
		})
	}
}

func TestDeleteVM(t *testing.T) {
	tests := []struct {
		name        string
		vmName      string
		statusCode  int
		expectError bool
	}{
		{
			name:        "successful delete",
			vmName:      "test-vm",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "not found",
			vmName:      "nonexistent",
			statusCode:  http.StatusNotFound,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, c := setupTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			})
			defer server.Close()

			err := c.DeleteVM(tt.vmName)
			if (err != nil) != tt.expectError {
				t.Errorf("DeleteVM() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
