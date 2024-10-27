package k3s

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// DeployMaster deploys k3s masters.
func DeployMaster(host, user, key string) (string, error) {
	out, err := exec.Command("curl", "-sfL", "https://get.k3s.io", "|", "INSTALL_K3S_EXEC='--write-kubeconfig-mode=755'", "sh", "-").CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return "", nil
}

// GetToken gets the k3s token.
func GetToken(host, user, key string) (string, error) {
	output, err := os.ReadFile("/var/lib/rancher/k3s/server/node-token")
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
	output, err := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return "", err
	}
	return string(output), nil
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
func UninstallMaster(host, user, key, password string) {
	out, err := exec.Command("/usr/local/bin/k3s-uninstall.sh", "||", "true").CombinedOutput()
	if err != nil {
		log.Println("error: %s, output: %s", err, out)
	}
	_, err = exec.Command("rm", "-rf", "/etc/rancher/k3s", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("rm", "-rf", "/var/lib/rancher", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("rm", "-rf", "/var/lib/rook", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("rm", "-rf", "/usr/local/bin/virtctl", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("rm", "-rf", "/etc/systemd/system/govnocloud2.service", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("rm", "-rf", "/etc/systemd/system/govnocloud2-web.service", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	_, err = exec.Command("systemctl", "daemon-reload", "||", "true").CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	err = os.Remove("/usr/local/bin/govnocloud2")
	if err != nil {
		log.Println(err)
	}
}
