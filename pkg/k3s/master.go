package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// MasterConfig holds configuration for k3s master operations
type MasterConfig struct {
	Host     string
	User     string
	Key      string
	Password string
	Timeout  int
	Retries  int
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
		Retries: 3,
	}
}

// DeployMaster deploys k3s master.
func DeployMaster(host, user, key string) error {
	cfg := NewMasterConfig(host, user, key)
	return cfg.Deploy()
}

// InstallK3sUp installs k3sup tool.
func (m *MasterConfig) InstallK3sUp() error {
	cmd := "curl -sLS https://get.k3sup.dev | sh && sudo install k3sup /usr/local/bin/"
	log.Println(cmd)
	out, err := ssh.Run(cmd, m.Host, m.Key, m.User, "", true, m.Timeout)
	log.Println(out)
	if err != nil {
		return fmt.Errorf("failed to install k3sup: %w", err)
	}
	return nil
}

// InstallK3sUp upgrades k3s up tool
func InstallK3sUp(host, user, key string) error {
	cfg := NewMasterConfig(host, user, key)
	return cfg.InstallK3sUp()
}

// Deploy deploys k3s master.
func (m *MasterConfig) Deploy() error {
	cmd := fmt.Sprintf("k3sup install --ip %s --user %s --ssh-key %s --sudo", m.Host, m.User, m.Key)
	log.Println(cmd)
	// retry 3 times
	var err error
	var out string
	for i := 0; i < m.Retries; i++ {
		out, err = ssh.Run(cmd, m.Host, m.Key, m.User, "", true, m.Timeout)
		log.Println(out)
		if err != nil {
			log.Println(err)
			log.Println("Retrying...")
		} else {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to deploy k3s master: %w", err)
	}
	cmd = fmt.Sprintf("sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config; sudo chown %s ~/.kube/config ; sudo chown %s /etc/rancher/k3s/k3s.yaml", m.User, m.User)
	out, err = ssh.Run(cmd, m.Host, m.Key, m.User, "", true, m.Timeout)
	log.Println(out)
	if err != nil {
		return fmt.Errorf("failed to copy k3s.yaml to ~/.kube/config: %w", err)
	}
	return nil
}

// UninstallMaster uninstalls k3s master.
func UninstallMaster(host, user, key, password string) error {
	cfg := NewMasterConfig(host, user, key)
	cfg.Password = password
	err := cfg.Uninstall()
	return err
}

// Uninstall removes k3s and related files from the master
func (m *MasterConfig) Uninstall() error {
	success := false
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
		{"sudo systemctl stop etcd.service", "stop etcd service"},
		{"sudo apt purge -y etcd-server etcd-client", "remove etcd"},
		{"sudo rm -rf /etc/systemd/system/govnocloud2-web.service || true", "remove web service file"},
		{"sudo rm -rf /etc/systemd/system/govnocloud2.service || true", "remove service file"},
		{"sudo systemctl daemon-reload || true", "reload systemd"},
		{"sudo rm -rf /usr/local/bin/govnocloud2 || true", "remove binary"},
		{"sudo rm -rf /usr/local/bin/govnocloud2-web || true", "remove web binary"},
		{"sudo rm -rf /var/www/govnocloud2 || true", "remove web dir"},
		{"sudo rm -rf /var/lib/etcd/* || true", "remove etcd data"},
	}
	for _, command := range cleanupCommands {
		out, err := ssh.Run(command.cmd, m.Host, m.Key, m.User, m.Password, false, 10)
		if err != nil {
			log.Printf("Failed to %s: %v\nOutput: %s", command.desc, err, out)
		} else {
			success = true
		}
	}
	if !success {
		return fmt.Errorf("failed to uninstall k3s master")
	}
	return nil
}
