package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// ListNodes retrieves a list of all nodes
func (c *Client) ListNodes() ([]string, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v0/nodes", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var nodes []string
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return nodes, nil
}

// GetNode retrieves details of a specific node
func (c *Client) GetNode(name string) (*types.Node, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/v0/nodes/%s", c.baseURL, name))
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var node types.Node
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// AddNode adds a new node to the cluster
func (c *Client) AddNode(node types.Node) error {
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v0/nodes", c.baseURL),
		"application/json",
		bytes.NewBuffer(nodeJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DeleteNode removes a node from the cluster
func (c *Client) DeleteNode(name string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/api/v0/nodes/%s", c.baseURL, name),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// RestartNode restarts a specific node
func (c *Client) RestartNode(name string) error {
	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/v0/nodes/%s/restart", c.baseURL, name),
		"application/json",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to restart node: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
