package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateVolume creates a new volume
func (c *Client) CreateVolume(name, namespace, size string) error {
	url := fmt.Sprintf("%s/volumes/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	volume := types.Volume{Name: name, Size: size}
	jsonBody, err := json.Marshal(volume)
	if err != nil {
		return fmt.Errorf("error marshalling volume: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error creating volume: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create volume: %d", resp.StatusCode)
	}
	return nil
}

// DeleteVolume deletes a volume
func (c *Client) DeleteVolume(name, namespace string) error {
	url := fmt.Sprintf("%s/volumes/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete volume: %d", resp.StatusCode)
	}

	return nil
}

// ListVolumes lists all volumes
func (c *Client) ListVolumes(namespace string) ([]string, error) {
	url := fmt.Sprintf("%s/volumes/%s", c.baseURL, namespace)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list volumes: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	volumes := []string{}
	err = json.Unmarshal(body, &volumes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return volumes, nil
}

// GetVolume gets details of a specific volume
func (c *Client) GetVolume(name, namespace string) (types.Volume, error) {
	url := fmt.Sprintf("%s/volumes/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return types.Volume{}, fmt.Errorf("error getting volume: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.Volume{}, fmt.Errorf("failed to get volume: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.Volume{}, fmt.Errorf("error reading response body: %w", err)
	}

	volume := types.Volume{}
	err = json.Unmarshal(body, &volume)
	if err != nil {
		return types.Volume{}, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return volume, nil
}
