package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// ListNodes retrieves a list of all nodes
func (c *Client) ListNodes() ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/nodes", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return nil, fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return nil, fmt.Errorf("failed to list nodes: %s %w", string(body), err)
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/nodes/%s", c.baseURL, name), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return nil, fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return nil, fmt.Errorf("failed to get node: %s %w", string(body), err)
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
func (c *Client) AddNode(name, host, masterHost, user, key string) error {
	node := types.Node{
		Name:       name,
		Host:       host,
		MasterHost: masterHost,
		User:       user,
		Key:        key,
	}
	if user == "" {
		node.User = "ubuntu"
	}
	if key == "" {
		node.Key = "/home/ubuntu/.ssh/id_rsa"
	}
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/nodes", c.baseURL), bytes.NewBuffer(nodeJSON))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return fmt.Errorf("failed to add node: %s %w", string(body), err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// DeleteNode removes a node from the cluster
func (c *Client) DeleteNode(name string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/nodes/%s", c.baseURL, name), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return fmt.Errorf("failed to delete node: %s %w", string(body), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// RestartNode restarts a specific node
func (c *Client) RestartNode(name string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/nodes/%s/restart", c.baseURL, name), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return fmt.Errorf("failed to restart node: %s %w", string(body), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// UpgradeNode upgrades a node
func (c *Client) UpgradeNode(ip string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/nodes/%s/upgrade", c.baseURL, ip), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err := c.SetAuthHeader(req); err != nil {
		return fmt.Errorf("error setting auth header: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		defer resp.Body.Close()
		return fmt.Errorf("failed to upgrade node: %s %w", string(body), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
