package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CreateNamespace creates a new namespace
func (c *Client) CreateNamespace(name string) error {
	url := fmt.Sprintf("%s/namespaces/%s", c.baseURL, name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error creating namespace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create namespace: %d", resp.StatusCode)
	}

	return nil
}

// DeleteNamespace deletes a namespace
func (c *Client) DeleteNamespace(name string) error {
	url := fmt.Sprintf("%s/namespaces/%s", c.baseURL, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting namespace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete namespace: %d", resp.StatusCode)
	}

	return nil
}

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces() ([]string, error) {
	url := fmt.Sprintf("%s/namespaces", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing namespaces: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list namespaces: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return strings.Split(string(body), " "), nil
}

// GetNamespace gets details of a specific namespace
func (c *Client) GetNamespace(name string) (string, error) {
	url := fmt.Sprintf("%s/namespaces/%s", c.baseURL, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting namespace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get namespace: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}
