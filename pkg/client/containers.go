package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateContainer creates a container.
func (c *Client) CreateContainer(name, image, namespace string, cpu, ram, disk, port int) error {
	container := types.Container{
		Name:      name,
		Image:     image,
		Namespace: namespace,
		CPU:       cpu,
		RAM:       ram,
		Disk:      disk,
		Port:      port,
	}

	data, err := json.Marshal(container)
	if err != nil {
		return fmt.Errorf("error marshaling container: %w", err)
	}

	url := fmt.Sprintf("%s/containers/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// set timeout to 600s
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

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
func (c *Client) ListContainers(namespace string) ([]types.Container, error) {
	url := fmt.Sprintf("%s/containers/%s", c.baseURL, namespace)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// set timeout to 600s
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error listing containers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing containers: status=%s body=%s", resp.Status, string(body))
	}
	bodyString, _ := io.ReadAll(resp.Body)
	log.Printf("resp: %+v", string(bodyString))
	var containers []types.Container
	err = json.Unmarshal(bodyString, &containers)
	if err != nil {
		return nil, fmt.Errorf("error decoding containers: %w", err)
	}

	return containers, nil
}

// GetContainer gets a container.
func (c *Client) GetContainer(name, namespace string) (types.Container, error) {
	if namespace == "" {
		return types.Container{}, fmt.Errorf("namespace is required")
	}
	url := fmt.Sprintf("%s/containers/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.Container{}, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// set timeout to 600s
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
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
func (c *Client) DeleteContainer(name, namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	url := fmt.Sprintf("%s/containers/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// set timeout to 600s
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

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
