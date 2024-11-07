package k3s

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployMaster deploys k3s masters.
func DeployMaster(host, user, key string) (string, error) {
	cmd := "curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC='--write-kubeconfig-mode=755' sh - || true"
	out, err := ssh.Run(cmd, host, key, user, "", true)
	if err != nil {
		return string(out), err
	}
	return "", nil
}

// GetToken gets the k3s token.
func GetToken(host, user, key string) (string, error) {
	output, err := ssh.Run("sudo cat /var/lib/rancher/k3s/server/node-token", host, key, user, "", false)
	if err != nil {
		return "", err
	}
	tokenSplit := strings.Split(string(output), ":")
	if len(tokenSplit) != 4 {
		return "", fmt.Errorf("invalid token")
	}
	res := tokenSplit[3][:len(tokenSplit[3])-1]
	return res, nil
}

// GetKubeconfig gets the k3s kubeconfig.
func GetKubeconfig(host, user, key string) (string, error) {
	output, err := ssh.Run("sudo cat /etc/rancher/k3s/k3s.yaml", host, key, user, "", false)
	if err != nil {
		return "", err
	}
	return output, nil
}

// WriteKubeconfig writes the k3s kubeconfig to the file.
func WriteKubeConfig(kubeconfig, path string) error {
	// Write the kubeconfig to the file
	// get dir from path
	dir := path[:strings.LastIndex(path, "/")]
	// create dir if not exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	err := os.WriteFile(path, []byte(kubeconfig), 0644)
	if err != nil {
		return err
	}
	return nil
}

// UninstallMaster uninstalls k3s master.
func UninstallMaster(host, user, key, password string) {
	out, err := ssh.Run("sudo /usr/local/bin/k3s-uninstall.sh || true", host, key, user, password, false)
	if err != nil {
		log.Println(out)
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /etc/rancher/k3s || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/rancher || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/kubelet || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/cni || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /var/lib/kubelet || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo systemctl stop govnocloud2.service", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo systemctl stop govnocloud2-web.service", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /etc/systemd/system/govnocloud2-web.service || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /etc/systemd/system/govnocloud2.service || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo systemctl daemon-reload || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
	_, err = ssh.Run("sudo rm -rf /usr/local/bin/govnocloud2 || true", host, key, user, password, false)
	if err != nil {
		log.Println(err)
	}
}
