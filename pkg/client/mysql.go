package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rusik69/govnocloud2/pkg/types"
)

// CreateMysql creates a mysql cluster.
func (c *Client) CreateMysql(name, namespace string, instances, routerInstances int) error {
	mysql := types.Mysql{
		Name:            name,
		Namespace:       namespace,
		Instances:       instances,
		RouterInstances: routerInstances,
	}

	data, err := json.Marshal(mysql)
	if err != nil {
		return fmt.Errorf("error marshaling mysql: %w", err)
	}

	url := fmt.Sprintf("%s/mysql/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error creating mysql cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating mysql cluster: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}

// GetMysql gets a mysql cluster.
func (c *Client) GetMysql(name, namespace string) (*types.Mysql, error) {
	url := fmt.Sprintf("%s/mysql/%s/%s", c.baseURL, namespace, name)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting mysql cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error getting mysql cluster: status=%s body=%s", resp.Status, string(body))
	}

	var mysql types.Mysql
	err = json.NewDecoder(resp.Body).Decode(&mysql)
	return &mysql, err
}

// ListMysql lists mysql clusters.
func (c *Client) ListMysql(namespace string) ([]types.Mysql, error) {
	url := fmt.Sprintf("%s/mysql/%s", c.baseURL, namespace)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error listing mysql clusters: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error listing mysql clusters: status=%s body=%s", resp.Status, string(body))
	}

	var mysqlClusters []types.Mysql
	err = json.NewDecoder(resp.Body).Decode(&mysqlClusters)
	return mysqlClusters, err
}

// DeleteMysql deletes a mysql cluster.
func (c *Client) DeleteMysql(name, namespace string) error {
	url := fmt.Sprintf("%s/mysql/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting mysql cluster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error deleting mysql cluster: status=%s body=%s", resp.Status, string(body))
	}

	return nil
}
