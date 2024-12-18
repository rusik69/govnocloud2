package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// NodeManager handles node operations
type NodeManager struct {
	kubectl KubectlRunner
	k3sup   K3supRunner
}

// KubectlRunner interface for executing kubectl commands
type KubectlRunner interface {
	Run(args ...string) ([]byte, error)
}

// K3supRunner interface for executing k3sup commands
type K3supRunner interface {
	Run(args ...string) ([]byte, error)
}

// DefaultKubectlRunner implements KubectlRunner using exec.Command
type DefaultKubectlRunner struct{}

func (k *DefaultKubectlRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("kubectl", args...).Output()
}

// DefaultK3supRunner implements K3supRunner using exec.Command
type DefaultK3supRunner struct{}

func (k *DefaultK3supRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("k3sup", args...).Output()
}

// NewNodeManager creates a new NodeManager instance
func NewNodeManager() *NodeManager {
	return &NodeManager{
		kubectl: &DefaultKubectlRunner{},
		k3sup:   &DefaultK3supRunner{},
	}
}

// ListNodesHandler handles HTTP requests to list nodes
func ListNodesHandler(c *gin.Context) {
	manager := NewNodeManager()
	nodes, err := manager.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to list nodes: %v", err),
		})
		return
	}
	log.Printf("nodes: %v", nodes)
	c.JSON(http.StatusOK, nodes)
}

// ListNodes returns a list of node names
func (m *NodeManager) ListNodes() ([]string, error) {
	out, err := m.kubectl.Run("get", "nodes", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		log.Printf("failed to get nodes: %v", err)
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	nodes := strings.Fields(string(out))
	if len(nodes) == 0 {
		log.Printf("no nodes found")
		return []string{}, nil
	}

	return nodes, nil
}

// GetNodeHandler handles HTTP requests to get node details
func GetNodeHandler(c *gin.Context) {
	nodeName := c.Param("name")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}

	manager := NewNodeManager()
	node, err := manager.GetNode(nodeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get node %s: %v", nodeName, err),
		})
		return
	}

	if node == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// GetNode retrieves details of a specific node
func (m *NodeManager) GetNode(name string) (*types.Node, error) {
	out, err := m.kubectl.Run("get", "node", name, "-o", "jsonpath='{.items[*].status.addresses[?(@.type==\"InternalIP\")].address}'")
	if err != nil {
		return nil, fmt.Errorf("failed to get node details: %w", err)
	}

	node := types.Node{
		Host:       string(out),
		User:       server.config.User,
		Key:        server.config.Key,
		Password:   server.config.Password,
		MasterHost: server.config.MasterHost,
	}
	return &node, nil
}

// DeleteNodeHandler handles HTTP requests to delete a node
func DeleteNodeHandler(c *gin.Context) {
	nodeName := c.Param("name")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}

	manager := NewNodeManager()
	if err := manager.DeleteNode(nodeName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to delete node %s: %v", nodeName, err),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteNode removes a node from the cluster
func (m *NodeManager) DeleteNode(name string) error {
	_, err := m.kubectl.Run("delete", "node", name)
	if err != nil {
		log.Printf("failed to delete node: %v", err)
		return fmt.Errorf("failed to delete node: %w", err)
	}
	return nil
}

// AddNodeHandler handles HTTP requests to add a node
func AddNodeHandler(c *gin.Context) {
	// This would typically involve generating a join token and returning instructions
	// Since this is not implemented, we return a proper status code and message
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var node types.Node
	if err := json.Unmarshal(body, &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
		return
	}

	manager := NewNodeManager()
	if err := manager.AddNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to add node: %v", err)})
		return
	}
}

// AddNode adds a node to the cluster
func (m *NodeManager) AddNode(node types.Node) error {
	out, err := m.k3sup.Run("join", "--ip", node.Host, "--server-ip", node.MasterHost, "--user", node.User, "--server-user", node.User, "--key", node.Key, "--k3s-extra-args", "--node-name", "node-"+node.Host)
	if err != nil {
		return fmt.Errorf("failed to add node: %s: %w", out, err)
	}
	return nil
}
