package k3s

import (
	"fmt"
	"log"

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
	}
}

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, password, master string) error {
	cfg := NewNodeConfig(host, user, key, password, master)
	return cfg.Deploy()
}

// Deploy installs k3s on the node
func (n *NodeConfig) Deploy() error {
	cmd := fmt.Sprintf("k3sup join --ip %s --user %s --ssh-key %s --sudo --server-ip %s --server-user %s", n.Host, n.User, n.Key, n.Master, n.User)
	log.Println(cmd)
	_, err := ssh.Run(cmd, n.Master, n.Key, n.User, n.Password, true, n.Timeout)
	if err != nil {
		return fmt.Errorf("failed to deploy k3s node: %w", err)
	}
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
