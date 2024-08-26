package k3s

import (
	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, master, token string) error {
	cmd := "curl -sfL https://get.k3s.io | K3S_URL=https://" + master + ":6443 K3S_TOKEN=" + token + " INSTALL_K3S_EXEC='--node-name=" + host + "' sh -"
	_, err := ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	return nil
}
