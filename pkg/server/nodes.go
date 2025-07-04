package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/govnocloud2/pkg/k8s"
	"github.com/rusik69/govnocloud2/pkg/ssh"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// NodeManager handles node operations
type NodeManager struct {
	kubectl KubectlRunner
}

// KubectlRunner interface for executing kubectl commands
type KubectlRunner interface {
	Run(args ...string) ([]byte, error)
}

// VirtctlRunner interface for executing virtctl commands
type VirtctlRunner interface {
	Run(args ...string) ([]byte, error)
}

// K3supRunner interface for executing k3sup commands
type K3supRunner interface {
	Run(args ...string) ([]byte, error)
}

// DefaultKubectlRunner implements KubectlRunner using exec.Command
type DefaultKubectlRunner struct{}

// DefaultVirtctlRunner implements VirtctlRunner using exec.Command
type DefaultVirtctlRunner struct{}

func (k *DefaultKubectlRunner) Run(args ...string) ([]byte, error) {
	log.Printf("running kubectl command: %v", args)
	return exec.Command("kubectl", args...).CombinedOutput()
}

func (k *DefaultVirtctlRunner) Run(args ...string) ([]byte, error) {
	log.Printf("running virtctl command: %v", args)
	cmd := exec.Command("virtctl", args...)
	cmd.Env = append(os.Environ(), "KUBECONFIG=/etc/rancher/k3s/k3s.yaml")
	return cmd.CombinedOutput()
}

// NewNodeManager creates a new NodeManager instance
func NewNodeManager() *NodeManager {
	return &NodeManager{
		kubectl: &DefaultKubectlRunner{},
	}
}

// ListNodesHandler handles HTTP requests to list nodes
func ListNodesHandler(c *gin.Context) {
	auth, _, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	nodes, err := nodeManager.ListNodes()
	if err != nil {
		log.Printf("failed to list nodes: %v", err)
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
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	nodeName := c.Param("name")
	if nodeName == "" {
		log.Printf("node name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	node, err := nodeManager.GetNode(nodeName)
	if err != nil {
		log.Printf("failed to get node %s: %v", nodeName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get node %s: %v", nodeName, err),
		})
		return
	}

	log.Printf("node: %+v", node)

	if node == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

// GetNode retrieves details of a specific node
func (m *NodeManager) GetNode(name string) (*types.Node, error) {
	// Get node IP
	ipOut, err := m.kubectl.Run("get", "node", name, "-o", "jsonpath={.status.addresses[?(@.type==\"InternalIP\")].address}")
	if err != nil {
		return nil, fmt.Errorf("failed to get node IP: %w", err)
	}

	// Get node status
	statusOut, err := m.kubectl.Run("get", "node", name, "-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}")
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}

	// Convert status to a more user-friendly format
	status := "Unknown"
	if strings.TrimSpace(string(statusOut)) == "True" {
		status = "Ready"
	} else if strings.TrimSpace(string(statusOut)) == "False" {
		status = "NotReady"
	}

	// Clean up the IP output
	host := strings.Trim(strings.TrimSpace(string(ipOut)), "'")
	if host == "" {
		return nil, fmt.Errorf("failed to get node IP address")
	}

	node := types.Node{
		Host:       host,
		User:       server.config.SSHUser,
		Key:        server.config.Key,
		Password:   server.config.SSHPassword,
		MasterHost: server.config.MasterHost,
		Status:     status,
	}
	return &node, nil
}

// DeleteNodeHandler handles HTTP requests to delete a node
func DeleteNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	nodeName := c.Param("name")
	if nodeName == "" {
		log.Printf("node name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	node, err := nodeManager.GetNode(nodeName)
	if err != nil {
		log.Printf("failed to get node %s: %v", nodeName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get node %s: %v", nodeName, err),
		})
	}

	if err := nodeManager.DeleteNode(node.Host); err != nil {
		log.Printf("failed to delete node %s: %v", nodeName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to delete node %s: %v", nodeName, err),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteNode removes a node from the cluster
func (m *NodeManager) DeleteNode(name string) error {
	cmd := "sudo /usr/local/bin/k3s-agent-uninstall.sh"
	out, err := ssh.Run(cmd, name, server.config.Key, server.config.SSHUser, "", true, 600)
	if err != nil {
		return fmt.Errorf("failed to uninstall k3s node: %w %s", err, out)
	}
	return nil
}

// AddNodeHandler handles HTTP requests to add a node
func AddNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	log.Println(string(body))
	var node types.Node
	if err := json.Unmarshal(body, &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse request body"})
		return
	}

	if err := nodeManager.AddNode(node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to add node: %v", err)})
		return
	}
}

// AddNode adds a node to the cluster
func (m *NodeManager) AddNode(node types.Node) error {
	err := k8s.DeployNode(node.Host, node.User, node.Key, node.Password, node.MasterHost)
	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}
	return nil
}

// RestartNodeHandler handles HTTP requests to restart a node
func RestartNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	nodeName := c.Param("name")
	if nodeName == "" {
		log.Printf("node name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}

	if err := nodeManager.RestartNode(nodeName); err != nil {
		log.Printf("failed to restart node %s: %v", nodeName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to restart node: %v", err)})
		return
	}
}

// RestartNode restarts a node
func (m *NodeManager) RestartNode(name string) error {
	node, err := m.GetNode(name)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	host := node.Host
	user := node.User
	key := node.Key
	password := node.Password
	if host == "" {
		return fmt.Errorf("host is required")
	}
	if user == "" {
		return fmt.Errorf("user is required")
	}
	if key == "" {
		return fmt.Errorf("key is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	_, err = m.kubectl.Run("cordon", "node", name)
	if err != nil {
		return fmt.Errorf("failed to cordon node: %w", err)
	}
	_, err = m.kubectl.Run("drain", "node", name, "--ignore-daemonsets", "--delete-emptydir-data")
	if err != nil {
		return fmt.Errorf("failed to drain node: %w", err)
	}
	rebootCmd := fmt.Sprintf("ssh -i %s %s@%s 'sudo reboot'", key, user, host)
	_, err = m.kubectl.Run(rebootCmd)
	if err != nil {
		return fmt.Errorf("failed to reboot node: %w", err)
	}
	_, err = m.kubectl.Run("uncordon", "node", name)
	if err != nil {
		return fmt.Errorf("failed to uncordon node: %w", err)
	}
	return nil
}

// SuspendNodeHandler handles HTTP requests to suspend a node
func SuspendNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	hostName := c.Param("name")
	if hostName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host name is required"})
		return
	}
	if err := nodeManager.SuspendNode(hostName, server.config.SSHUser, server.config.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to suspend node: %v", err)})
		return
	}
}

// SuspendNode suspends a node
func (m *NodeManager) SuspendNode(host, user, key string) error {
	cmd := fmt.Sprintf("ssh -i %s %s@%s 'sudo systemctl suspend'", key, user, host)
	_, err := ssh.Run(cmd, host, key, user, "", true, 10)
	if err != nil {
		return fmt.Errorf("failed to suspend node: %w", err)
	}
	return nil
}

// ResumeNodeHandler handles HTTP requests to resume a node
func ResumeNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	hostName := c.Param("name")
	if hostName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "host name is required"})
		return
	}
	if err := nodeManager.ResumeNode(hostName, server.config.SSHUser, server.config.Key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to resume node: %v", err)})
		return
	}
}

// ResumeNode resumes a node
func (m *NodeManager) ResumeNode(host, user, key string) error {
	return nil
}

// UpgradeNodeHandler handles HTTP requests to upgrade a node
func UpgradeNodeHandler(c *gin.Context) {
	auth, username, err := CheckAuth(c)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, fmt.Sprintf("failed to check auth: %v", err))
		return
	}
	if !auth {
		respondWithError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !CheckAdminAccess(username) {
		respondWithError(c, http.StatusForbidden, "user does not have admin access")
		return
	}
	hostName := c.Param("name")
	if hostName == "" {
		log.Printf("host name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "host name is required"})
		return
	}
	node, err := nodeManager.GetNode(hostName)
	if err != nil {
		log.Printf("failed to get node %s: %v", hostName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to get node: %v", err)})
		return
	}
	if err := nodeManager.UpgradeNode(node.Host, node.User, node.Key); err != nil {
		log.Printf("failed to upgrade node %s@%s:%s %v", node.User, node.Host, node.Key, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to upgrade node: %v", err)})
		return
	}
}

// UpgradeNode upgrades a node
func (m *NodeManager) UpgradeNode(host, user, key string) error {
	cmd := "sudo apt-get update && sudo apt-get upgrade -y"
	log.Printf("upgrading node %s with command %s", host, cmd)
	out, err := ssh.Run(cmd, host, key, user, "", false, 600)
	if err != nil {
		return fmt.Errorf("failed to upgrade node: %w", err)
	}
	log.Printf("upgrade node %s output: %s", host, out)
	return nil
}
