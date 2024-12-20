package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// Client represents an HTTP client for VM operations
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new VM client
func NewClient(host, port string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%s/api/v0", host, port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateVM creates a VM.
func (c *Client) CreateVM(name, image, size, namespace string) error {
	vm := types.VM{
		Name:      name,
		Image:     image,
		Size:      size,
		Namespace: namespace,
	}

	data, err := json.Marshal(vm)
	if err != nil {
		return fmt.Errorf("error marshaling VM: %w", err)
	}

	url := fmt.Sprintf("%s/vms/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating VM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating VM: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}

// ListVMs lists VMs.
func (c *Client) ListVMs(namespace string) ([]types.VM, error) {
	url := fmt.Sprintf("%s/vms/%s", c.baseURL, namespace)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing VMs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing VMs: status=%s body=%s", resp.Status, string(body))
	}

	var vms []types.VM
	if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
		return nil, fmt.Errorf("error decoding VMs: %w", err)
	}

	return vms, nil
}

// GetVM gets a VM.
func (c *Client) GetVM(name, namespace string) (*types.VM, error) {
	url := fmt.Sprintf("%s/vms/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting VM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting VM: status=%s body=%s", resp.Status, string(body))
	}

	var vm types.VM
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return nil, fmt.Errorf("error decoding VM: %w", err)
	}

	return &vm, nil
}

// DeleteVM deletes a VM.
func (c *Client) DeleteVM(name, namespace string) error {
	url := fmt.Sprintf("%s/vms/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting VM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting VM: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}
