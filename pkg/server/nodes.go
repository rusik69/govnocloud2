package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
)

// NodeManager handles node operations
type NodeManager struct {
	kubectl KubectlRunner
}

// KubectlRunner interface for executing kubectl commands
type KubectlRunner interface {
	Run(args ...string) ([]byte, error)
}

// DefaultKubectlRunner implements KubectlRunner using exec.Command
type DefaultKubectlRunner struct{}

func (k *DefaultKubectlRunner) Run(args ...string) ([]byte, error) {
	return exec.Command("kubectl", args...).Output()
}

// NewNodeManager creates a new NodeManager instance
func NewNodeManager() *NodeManager {
	return &NodeManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// Node represents a Kubernetes node
type Node struct {
	Name       string            `json:"name"`
	Status     string            `json:"status,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Conditions []NodeCondition   `json:"conditions,omitempty"`
}

// NodeCondition represents the condition of a node
type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
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
func (m *NodeManager) GetNode(name string) (*Node, error) {
	out, err := m.kubectl.Run("get", "node", name, "-o", "json")
	if err != nil {
		log.Printf("failed to get node details: %v", err)
		return nil, fmt.Errorf("failed to get node details: %w", err)
	}

	var node Node
	if err := json.Unmarshal(out, &node); err != nil {
		log.Printf("failed to parse node details: %v", err)
		return nil, fmt.Errorf("failed to parse node details: %w", err)
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
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "node addition via API is not implemented - please use kubeadm join command",
	})
}
