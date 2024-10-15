package k3s

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployMaster deploys k3s masters.
func DeployMaster(host, user, key string) error {
	cmd := "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--write-kubeconfig-mode=755' sh -"
	log.Println(cmd)
	_, err := ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	return nil
}

// GetToken gets the k3s token.
func GetToken(host, user, key string) (string, error) {
	cmd := "sudo cat /var/lib/rancher/k3s/server/node-token"
	output, err := ssh.Run(cmd, host, key, user, false)
	if err != nil {
		return "", err
	}
	tokenSplit := strings.Split(output, ":")
	if len(tokenSplit) != 4 {
		return "", fmt.Errorf("invalid token")
	}
	res := tokenSplit[3][:len(tokenSplit[3])-1]
	return res, nil
}

// GetKubeconfig gets the k3s kubeconfig.
func GetKubeconfig(host, user, key string) (string, error) {
	cmd := "sudo cat /etc/rancher/k3s/k3s.yaml"
	output, err := ssh.Run(cmd, host, key, user, false)
	if err != nil {
		return "", err
	}
	return output, nil
}

// WriteKubeconfig writes the k3s kubeconfig to the file.
func WriteKubeConfig(kubeconfig, path string) error {
	// Write the kubeconfig to the file
	err := os.WriteFile(path, []byte(kubeconfig), 0644)
	if err != nil {
		return err
	}
	return nil
}

// UninstallMaster uninstalls k3s master.
func UninstallMaster(host, user, key string) error {
	cmd := "sudo /usr/local/bin/k3s-uninstall.sh || true"
	_, err := ssh.Run(cmd, host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /etc/rancher/k3s || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/rancher || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/rook || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /usr/local/bin/virtctl || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /etc/systemd/system/govnocloud2.service || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo rm -rf /etc/systemd/system/govnocloud2-web.service || true", host, key, user, true)
	if err != nil {
		return err
	}
	_, err = ssh.Run("sudo systemctl daemon-reload || true", host, key, user, true)
	if err != nil {
		return err
	}
	return nil
}
