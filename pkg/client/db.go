package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateDB creates a database.
func (c *Client) CreateDB(name, namespace, dbType, dbSize string) error {
	db := types.DB{
		Name:      name,
		Namespace: namespace,
		Type:      dbType,
		Size:      dbSize,
	}

	data, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("error marshaling database: %w", err)
	}

	url := fmt.Sprintf("%s/dbs/%s", c.baseURL, namespace)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
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

// GetDB gets a database.
func (c *Client) GetDB(name, namespace string) (types.DB, error) {
	url := fmt.Sprintf("%s/dbs/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return types.DB{}, fmt.Errorf("error getting database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.DB{}, fmt.Errorf("error getting database: status=%s body=%s", resp.Status, string(body))
	}

	var db types.DB
	err = json.NewDecoder(resp.Body).Decode(&db)
	return db, err
}

// ListDBs lists databases.
func (c *Client) ListDBs(namespace string) ([]types.DB, error) {
	url := fmt.Sprintf("%s/dbs/%s", c.baseURL, namespace)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing databases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing databases: status=%s body=%s", resp.Status, string(body))
	}

	var dbs []types.DB
	err = json.NewDecoder(resp.Body).Decode(&dbs)
	return dbs, err
}

// DeleteDB deletes a database.
func (c *Client) DeleteDB(name, namespace string) error {
	url := fmt.Sprintf("%s/dbs/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error deleting database: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting database: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}
