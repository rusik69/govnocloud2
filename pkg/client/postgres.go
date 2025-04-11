package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreatePostgres creates a postgres cluster.
func (c *Client) CreatePostgres(name, namespace, size string, replicas int, storage int) error {
	db := types.Postgres{
		Name:      name,
		Namespace: namespace,
		Size:      size,
		Replicas:  replicas,
		Storage:   storage,
	}

	data, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("error marshaling database: %w", err)
	}

	url := fmt.Sprintf("%s/postgres/%s/%s", c.baseURL, namespace, name)

	// set timeout to 600s
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error creating database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating database: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}

// GetPostgres gets a postgres cluster.
func (c *Client) GetPostgres(name, namespace string) (types.Postgres, error) {
	url := fmt.Sprintf("%s/postgres/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.Postgres{}, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return types.Postgres{}, fmt.Errorf("error getting postgres cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.Postgres{}, fmt.Errorf("error getting postgres cluster: status=%s body=%s", resp.Status, string(body))
	}

	var db types.Postgres
	err = json.NewDecoder(resp.Body).Decode(&db)
	return db, err
}

// ListPostgres lists postgres clusters.
func (c *Client) ListPostgres(namespace string) ([]types.Postgres, error) {
	url := fmt.Sprintf("%s/postgres/%s", c.baseURL, namespace)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing databases: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error listing databases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing postgres clusters: status=%s body=%s", resp.Status, string(body))
	}

	var dbs []types.Postgres
	err = json.NewDecoder(resp.Body).Decode(&dbs)
	return dbs, err
}

// DeletePostgres deletes a postgres cluster.
func (c *Client) DeletePostgres(name, namespace string) error {
	url := fmt.Sprintf("%s/postgres/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error deleting postgres cluster: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User", c.username)
	req.Header.Set("Password", c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting postgres cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting postgres cluster: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}
