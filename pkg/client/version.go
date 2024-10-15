package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetServerVersion returns the server version.
func GetServerVersion(host, port string) (string, error) {
	url := fmt.Sprintf("http://%s:%s/version", host, port)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error getting server version: %w", err)
	}
	defer resp.Body.Close()
	var ver struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ver); err != nil {
		return "", fmt.Errorf("error decoding server version: %w", err)
	}
	return ver.Version, nil
}
