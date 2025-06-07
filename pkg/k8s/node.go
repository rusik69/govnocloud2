package k8s

import (
	"fmt"
	"log"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// NodeConfig holds configuration for k3s node operations
type NodeConfig struct {
	Host     string
	User     string
	Key      string
	Password string
	Master   string
	Timeout  int
	Retry    int
}

// NewNodeConfig creates a new node configuration with defaults
func NewNodeConfig(host, user, key, password, master string) *NodeConfig {
	return &NodeConfig{
		Host:     host,
		User:     user,
		Key:      key,
		Password: password,
		Master:   master,
		Timeout:  600,
		Retry:    3,
	}
}

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, password, master string) error {
	cfg := NewNodeConfig(host, user, key, password, master)
	return cfg.Deploy()
}

// generateNodeName generates a node name from the host
func (n *NodeConfig) generateNodeName() string {
	// Replace dots with dashes to create valid node name
	nodeName := strings.ReplaceAll(n.Host, ".", "-")
	// Add worker prefix
	return fmt.Sprintf("node-%s", nodeName)
}

// Deploy installs k3s on the node
func (n *NodeConfig) Deploy() error {
	nodeName := n.generateNodeName()
	cmd := fmt.Sprintf(
		"k3sup join --ip %s --user %s --ssh-key %s --server-ip %s --server-user %s --k3s-extra-args '--node-name %s' --sudo",
		n.Host,
		n.User,
		n.Key,
		n.Master,
		n.User,
		nodeName,
	)
	log.Printf("Running: %s", cmd)
	var err error
	var out string
	// retry 3 times
	for i := 0; i < n.Retry; i++ {
		out, err = ssh.Run(cmd, n.Master, n.Key, n.User, n.Password, true, n.Timeout)
		if err != nil {
			log.Println(err)
			log.Println("Retrying...")
		} else {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to deploy k3s node: %s: %w", out, err)
	}
	log.Printf("Node deployment output: %s", out)
	return nil
}

// UninstallNode uninstalls k3s node.
func UninstallNode(master, host, user, key, password string) error {
	cfg := NewNodeConfig(host, user, key, password, master)
	return cfg.Uninstall()
}

// Uninstall removes k3s and related files from the node
func (n *NodeConfig) Uninstall() error {
	cleanupCommands := []struct {
		cmd  string
		desc string
	}{
		{
			cmd:  fmt.Sprintf("ssh %s@%s 'sudo /usr/local/bin/k3s-agent-uninstall.sh || true'", n.User, n.Host),
			desc: "uninstall k3s agent",
		},
		{
			cmd:  fmt.Sprintf("ssh %s@%s 'sudo rm -rf /etc/rancher || true'", n.User, n.Host),
			desc: "remove k3s config",
		},
		{
			cmd:  fmt.Sprintf("ssh %s@%s 'sudo rm -rf /var/lib/rancher || true'", n.User, n.Host),
			desc: "remove rancher data",
		},
		{
			cmd:  fmt.Sprintf("ssh %s@%s 'sudo rm -rf /var/lib/kubelet || true'", n.User, n.Host),
			desc: "remove kubelet data",
		},
	}

	for _, command := range cleanupCommands {
		log.Printf("Running: %s", command.cmd)
		out, err := ssh.Run(command.cmd, n.Master, n.Key, n.User, n.Password, true, 10)
		if err != nil {
			return fmt.Errorf("failed to %s: %s: %w", command.desc, out, err)
		}
	}

	return nil
}
