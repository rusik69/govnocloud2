package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateClickhouse creates a clickhouse cluster.
func (c *Client) CreateClickhouse(name, namespace string, replicas int) error {
	clickhouse := types.Clickhouse{
		Name:      name,
		Namespace: namespace,
		Replicas:  replicas,
	}

	data, err := json.Marshal(clickhouse)
	if err != nil {
		return fmt.Errorf("error marshaling clickhouse: %w", err)
	}
	// Create a new request to properly set headers
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/clickhouse/%s/%s", c.baseURL, namespace, name), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set content type and authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Use the client's httpClient to send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error creating clickhouse cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating clickhouse cluster: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}

// GetClickhouse gets a clickhouse cluster.
func (c *Client) GetClickhouse(name, namespace string) (*types.Clickhouse, error) {
	// Create a new request to properly set headers
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clickhouse/%s/%s", c.baseURL, namespace, name), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set content type and authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Use the client's httpClient to send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting clickhouse cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting clickhouse cluster: status=%s body=%s", resp.Status, string(body))
	}

	var clickhouse types.Clickhouse
	err = json.NewDecoder(resp.Body).Decode(&clickhouse)
	return &clickhouse, err
}

// ListClickhouse lists clickhouse clusters.
func (c *Client) ListClickhouse(namespace string) ([]types.Clickhouse, error) {
	// Create a new request to properly set headers
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/clickhouse/%s", c.baseURL, namespace), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set content type and authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Use the client's httpClient to send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error listing clickhouse clusters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing clickhouse clusters: status=%s body=%s", resp.Status, string(body))
	}

	var clickhouseClusters []types.Clickhouse
	err = json.NewDecoder(resp.Body).Decode(&clickhouseClusters)
	return clickhouseClusters, err
}

// DeleteClickhouse deletes a clickhouse cluster.
func (c *Client) DeleteClickhouse(name, namespace string) error {
	// Create a new request to properly set headers
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/clickhouse/%s/%s", c.baseURL, namespace, name), nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	// Set content type and authentication headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	// Use the client's httpClient to send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting clickhouse cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting clickhouse cluster: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}
