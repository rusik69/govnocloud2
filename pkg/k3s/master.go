package k3s

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// MasterConfig holds configuration for k3s master operations
type MasterConfig struct {
	Host     string
	User     string
	Key      string
	Password string
	Timeout  int
}

// K3sFiles represents important k3s file paths
type K3sFiles struct {
	NodeToken  string
	Kubeconfig string
}

// NewMasterConfig creates a new master configuration with defaults
func NewMasterConfig(host, user, key string) *MasterConfig {
	return &MasterConfig{
		Host:    host,
		User:    user,
		Key:     key,
		Timeout: 600,
	}
}

// DeployMaster deploys k3s master.
func DeployMaster(host, user, key string) error {
	cfg := NewMasterConfig(host, user, key)
	return cfg.Deploy()
}

// Deploy installs k3s on the master node
func (m *MasterConfig) Deploy() error {
	cmd := "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--write-kubeconfig-mode=755' sh -"
	log.Println(cmd)
	out, err := ssh.Run(cmd, m.Host, m.Key, m.User, "", true, m.Timeout)
	log.Println(out)
	if err != nil {
		return fmt.Errorf("failed to deploy k3s master: %w", err)
	}
	return nil
}

// GetToken gets the k3s token.
func GetToken(host, user, key string) (string, error) {
	cfg := NewMasterConfig(host, user, key)
	return cfg.GetToken()
}

// GetToken retrieves the node token from the master
func (m *MasterConfig) GetToken() (string, error) {
	output, err := ssh.Run("sudo cat /var/lib/rancher/k3s/server/node-token", m.Host, m.Key, m.User, "", false, 5)
	if err != nil {
		return "", fmt.Errorf("failed to get node token: %w", err)
	}

	tokenSplit := strings.Split(string(output), ":")
	if len(tokenSplit) != 4 {
		return "", fmt.Errorf("invalid token format")
	}

	return strings.TrimSpace(tokenSplit[3]), nil
}

// GetKubeconfig gets the k3s kubeconfig.
func GetKubeconfig(host, user, key string) (string, error) {
	cfg := NewMasterConfig(host, user, key)
	return cfg.GetKubeconfig()
}

// GetKubeconfig retrieves the kubeconfig from the master
func (m *MasterConfig) GetKubeconfig() (string, error) {
	output, err := ssh.Run("sudo cat /etc/rancher/k3s/k3s.yaml", m.Host, m.Key, m.User, "", false, 5)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	return output, nil
}

// WriteKubeconfig writes the kubeconfig to a file
func WriteKubeConfig(kubeconfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create kubeconfig directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(kubeconfig), 0644); err != nil {
		return fmt.Errorf("failed to write kubeconfig file: %w", err)
	}

	return nil
}

// UninstallMaster uninstalls k3s master.
func UninstallMaster(host, user, key, password string) {
	cfg := NewMasterConfig(host, user, key)
	cfg.Password = password
	cfg.Uninstall()
}

// Uninstall removes k3s and related files from the master
func (m *MasterConfig) Uninstall() {
	cleanupCommands := []struct {
		cmd  string
		desc string
	}{
		{"sudo /usr/local/bin/k3s-uninstall.sh || true", "uninstall k3s"},
		{"sudo rm -rf /etc/rancher/k3s || true", "remove k3s config"},
		{"sudo rm -rf /var/lib/rancher || true", "remove rancher data"},
		{"sudo rm -rf /var/lib/kubelet || true", "remove kubelet data"},
		{"sudo rm -rf /var/lib/cni || true", "remove cni data"},
		{"sudo systemctl stop govnocloud2.service", "stop govnocloud2 service"},
		{"sudo systemctl stop govnocloud2-web.service", "stop govnocloud2-web service"},
		{"sudo rm -rf /etc/systemd/system/govnocloud2-web.service || true", "remove web service file"},
		{"sudo rm -rf /etc/systemd/system/govnocloud2.service || true", "remove service file"},
		{"sudo systemctl daemon-reload || true", "reload systemd"},
		{"sudo rm -rf /usr/local/bin/govnocloud2 || true", "remove binary"},
	}

	for _, command := range cleanupCommands {
		out, err := ssh.Run(command.cmd, m.Host, m.Key, m.User, m.Password, false, 10)
		if err != nil {
			log.Printf("Failed to %s: %v\nOutput: %s", command.desc, err, out)
		}
	}
}
