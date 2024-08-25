package k3s

import (
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployMaster deploys k3s masters.
func DeployMaster(host, user, key string) (string, error) {
	cmd := "curl -sfL https://get.k3s.io | sh -"
	log.Println(cmd)
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	return output, nil
}

// GetToken gets the k3s token.
func GetToken(host, user, key string) (string, error) {
	cmd := "sudo cat /var/lib/rancher/k3s/server/node-token"
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	return output, nil
}

// GetKubeconfig gets the k3s kubeconfig.
func GetKubeconfig(host, user, key string) (string, error) {
	cmd := "sudo cat /etc/rancher/k3s/k3s.yaml"
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	return output, nil
}

// UninstallMaster uninstalls k3s master.
func UninstallMaster(host, user, key string) (string, error) {
	cmd := "sudo /usr/local/bin/k3s-uninstall.sh || true"
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	output2, err := ssh.Run("sudo rm -rf /etc/rancher/k3s || true", host, key, user)
	if err != nil {
		return "", err
	}
	output3, err := ssh.Run("sudo rm -rf /var/lib/rancher || true", host, key, user)
	if err != nil {
		return "", err
	}
	return output + output2 + output3, nil
}

// UninstallNode uninstalls k3s node.
func UninstallNode(host, user, key string) (string, error) {
	cmd := "sudo /usr/local/bin/k3s-agent-uninstall.sh || true"
	output, err := ssh.Run(cmd, host, key, user)
	if err != nil {
		return "", err
	}
	output2, err := ssh.Run("sudo rm -rf /etc/rancher/k3s || true", host, key, user)
	if err != nil {
		return "", err
	}
	output3, err := ssh.Run("sudo rm -rf /var/lib/rancher || true", host, key, user)
	if err != nil {
		return "", err
	}
	return output + output2 + output3, nil
}
