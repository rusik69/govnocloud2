package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// ListUsers returns a list of all users
func (c *Client) ListUsers() ([]types.User, error) {
	url := fmt.Sprintf("%s/users", c.baseURL)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list users: server returned %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data  []types.User `json:"data"`
		Error string       `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != "" {
		return nil, fmt.Errorf("server error: %s", response.Error)
	}

	return response.Data, nil
}

// GetUser gets a user by name
func (c *Client) GetUser(name string) (types.User, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, name)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return types.User{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return types.User{}, fmt.Errorf("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.User{}, fmt.Errorf("failed to get user: server returned %d: %s", resp.StatusCode, string(body))
	}

	var response types.User
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return types.User{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

// CreateUser creates a new user
func (c *Client) CreateUser(name string, user types.User) error {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, name)
	user.Name = name // Ensure consistency

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create user: server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteUser deletes a user by name
func (c *Client) DeleteUser(name string) error {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user: server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetUserPassword sets a user's password
func (c *Client) SetUserPassword(name, password string) error {
	url := fmt.Sprintf("%s/users/%s/password", c.baseURL, name)
	data, err := json.Marshal(password)
	if err != nil {
		return fmt.Errorf("failed to marshal password: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to set user password: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set user password: server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// AddNamespaceToUser adds a namespace to a user's list of accessible namespaces
func (c *Client) AddNamespaceToUser(name, namespace string) error {
	url := fmt.Sprintf("%s/users/%s/namespaces/%s", c.baseURL, name, namespace)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add namespace to user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add namespace to user: server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemoveNamespaceFromUser removes a namespace from a user's list of accessible namespaces
func (c *Client) RemoveNamespaceFromUser(name, namespace string) error {
	url := fmt.Sprintf("%s/users/%s/namespaces/%s", c.baseURL, name, namespace)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to remove namespace from user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove namespace from user: server returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
