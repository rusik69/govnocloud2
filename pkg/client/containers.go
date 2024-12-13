package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateContainer creates a container.
func (c *Client) CreateContainer(name, image, namespace string) error {
	container := types.Container{
		Name:      name,
		Image:     image,
		Namespace: namespace,
	}

	data, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("error marshaling container: %w", err)
	}

	url := fmt.Sprintf("%s/containers", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error creating container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating container: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}

// ListContainers lists containers.
func (c *Client) ListContainers() ([]types.Container, error) {
	url := fmt.Sprintf("%s/containers", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing containers: status=%s body=%s", resp.Status, string(body))
	}

	var containers []types.Container
	err = json.NewDecoder(resp.Body).Decode(&containers)
	if err != nil {
		return nil, fmt.Errorf("error decoding containers: %w", err)
	}

	return containers, nil
}

// GetContainer gets a container.
func (c *Client) GetContainer(name string) (types.Container, error) {
	url := fmt.Sprintf("%s/containers/%s", c.baseURL, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return types.Container{}, fmt.Errorf("error getting container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.Container{}, fmt.Errorf("error getting container: status=%s body=%s", resp.Status, string(body))
	}

	var container types.Container
	err = json.NewDecoder(resp.Body).Decode(&container)
	if err != nil {
		return types.Container{}, fmt.Errorf("error decoding container: %w", err)
	}

	return container, nil
}

// DeleteContainer deletes a container.
func (c *Client) DeleteContainer(name string) error {
	url := fmt.Sprintf("%s/containers/%s", c.baseURL, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting container: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting container: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}