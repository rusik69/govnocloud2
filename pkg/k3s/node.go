package k3s

import "github.com/rusik69/govnocloud2/pkg/ssh"

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, master, token string) (string, error) {
	cmd := "curl -sfL https://get.k3s.io | K3S_URL=https://" + master + ":6443 K3S_TOKEN=" + token + " sh -"
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	return output, nil
}
