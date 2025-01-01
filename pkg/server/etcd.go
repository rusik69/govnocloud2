package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rusik69/govnocloud2/pkg/types"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient represents an etcd client
type EtcdClient struct {
	client *clientv3.Client
}

// NewEtcdClient creates a new etcd client
func NewEtcdClient(host, port string) (*EtcdClient, error) {
	endpoints := []string{fmt.Sprintf("%s:%s", host, port)}
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdClient{
		client: client,
	}, nil
}

// Put stores a key-value pair in etcd
func (e *EtcdClient) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}

	log.Printf("Successfully stored key %s", key)
	return nil
}

// Get retrieves a value from etcd by key
func (e *EtcdClient) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key %s not found", key)
	}

	return string(resp.Kvs[0].Value), nil
}

// List retrieves all key-value pairs with a given prefix
func (e *EtcdClient) List(prefix string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to list keys with prefix %s: %w", prefix, err)
	}

	result := make(map[string]string)
	for _, kv := range resp.Kvs {
		result[string(kv.Key)] = string(kv.Value)
	}

	return result, nil
}

// Close closes the etcd client connection
func (e *EtcdClient) Close() error {
	if err := e.client.Close(); err != nil {
		return fmt.Errorf("failed to close etcd client: %w", err)
	}
	return nil
}

// Example usage functions:

// StoreNode stores node information in etcd
func (e *EtcdClient) StoreNode(node types.Node) error {
	key := fmt.Sprintf("/nodes/%s", node.Name)
	nodeJson, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}
	return e.Put(key, string(nodeJson))
}

// GetNode retrieves node information from etcd
func (e *EtcdClient) GetNode(nodeName string) (types.Node, error) {
	key := fmt.Sprintf("/nodes/%s", nodeName)
	nodeJson, err := e.Get(key)
	if err != nil {
		return types.Node{}, fmt.Errorf("failed to get node: %w", err)
	}
	var node types.Node
	err = json.Unmarshal([]byte(nodeJson), &node)
	if err != nil {
		return types.Node{}, fmt.Errorf("failed to unmarshal node: %w", err)
	}
	return node, nil
}

// ListNodes retrieves all nodes from etcd
func (e *EtcdClient) ListNodes() ([]types.Node, error) {
	nodesJson, err := e.List("/nodes/")
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	var nodes []types.Node
	for _, nodeJson := range nodesJson {
		var node types.Node
		err = json.Unmarshal([]byte(nodeJson), &node)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal node: %w", err)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}
