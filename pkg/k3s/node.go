package k3s

import (
	"errors"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// DeployNode deploys k3s nodes.
func DeployNode(host, user, key, password, master, token string) (string, error) {
	cmd := "curl -sfL https://get.k3s.io | K3S_URL=https://" + master + ":6443 K3S_TOKEN=" + token + " INSTALL_K3S_EXEC='--node-name=" + host + "' sh -"
	log.Println(cmd)
	out, err := ssh.Run(cmd, host, key, user, password, false, 600)
	if err != nil {
		return "", err
	}
	return out, nil
}

// UninstallNode uninstalls k3s node.
func UninstallNode(master, host, user, key, password string) error {
	cmd := "ssh " + user + "@" + host + " 'sudo /usr/local/bin/k3s-agent-uninstall.sh || true'"
	log.Println(cmd)
	out, err := ssh.Run(cmd, master, key, user, password, true, 600)
	if err != nil {
		return errors.New(string(out))
	}
	cmd = "ssh " + user + "@" + host + " 'sudo rm -rf /etc/rancher/k3s || true'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, master, key, user, password, true, 10)
	if err != nil {
		return errors.New(string(out))
	}
	cmd = "ssh " + user + "@" + host + " 'sudo rm -rf /var/lib/rancher || true'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, master, key, user, password, true, 10)
	if err != nil {
		return errors.New(string(out))
	}
	cmd = "ssh " + user + "@" + host + " 'sudo rm -rf /var/lib/kubelet || true'"
	log.Println(cmd)
	out, err = ssh.Run(cmd, master, key, user, password, true, 10)
	if err != nil {
		return errors.New(string(out))
	}
	return nil
}
