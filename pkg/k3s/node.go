package k3s

import (
	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, password, master, token string) (string, error) {
	cmd := "curl -sfL https://get.k3s.io | K3S_URL=https://" + master + ":6443 K3S_TOKEN=" + token + " INSTALL_K3S_EXEC='--node-name=" + host + "' sh -"
	out, err := ssh.Run(cmd, host, key, user, password, false)
	if err != nil {
		return "", err
	}
	return out, nil
}

// UninstallNode uninstalls k3s node.
func UninstallNode(host, user, key, password string) error {
	cmd := "sudo /usr/local/bin/k3s-agent-uninstall.sh || true"
	_, err := ssh.Run(cmd, host, key, user, password, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /etc/rancher/k3s || true", host, key, user, password, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/rancher || true", host, key, user, password, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/rook || true", host, key, user, password, true)
	if err != nil {
		return err
	}
	return nil
}
