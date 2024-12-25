package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// VersionResponse represents the server version response
type VersionResponse struct {
	Version string `json:"version"`
}

// GetServerVersion returns the server version.
func GetServerVersion(host, port string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("http://%s:%s/version", host, port)

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting server version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ver VersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&ver); err != nil {
		return "", fmt.Errorf("error decoding server version: %w", err)
	}

	return ver.Version, nil
}
