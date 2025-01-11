package k3s

import (
	"fmt"
	"log"
	"time"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

type KubeVirtConfig struct {
	Host    string
	User    string
	Key     string
	Version string
}

func NewKubeVirtConfig(host, user, key string) *KubeVirtConfig {
	return &KubeVirtConfig{
		Host:    host,
		User:    user,
		Key:     key,
		Version: "v1.4.0",
	}
}

func InstallKubeVirt(host, user, key string) error {
	cfg := NewKubeVirtConfig(host, user, key)
	baseURL := fmt.Sprintf("https://github.com/kubevirt/kubevirt/releases/download/%s", cfg.Version)

	// Install operator and CR
	manifests := []string{"kubevirt-operator.yaml", "kubevirt-cr.yaml"}
	for _, manifest := range manifests {
		cmd := fmt.Sprintf("kubectl apply -f %s/%s --wait=true --timeout=300s", baseURL, manifest)
		log.Println(cmd)
		if out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60); err != nil {
			return fmt.Errorf("failed to apply %s: %w", manifest, err)
		} else {
			log.Println(out)
		}
	}

	// Install virtctl
	virtctlCmd := fmt.Sprintf("sudo curl -L -o /usr/local/bin/virtctl %s/virtctl-%s-linux-amd64 && sudo chmod +x /usr/local/bin/virtctl",
		baseURL, cfg.Version)
	log.Println(virtctlCmd)
	if out, err := ssh.Run(virtctlCmd, cfg.Host, cfg.Key, cfg.User, "", true, 60); err != nil {
		return fmt.Errorf("failed to install virtctl: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for KubeVirt to be ready
	time.Sleep(5 * time.Second)
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-operator -n kubevirt"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, cfg.Host, cfg.Key, cfg.User, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt: %w", err)
	}

	return nil
}

func InstallKubeVirtManager(host, user, key string) error {
	managerURL := "https://raw.githubusercontent.com/kubevirt-manager/kubevirt-manager/main/kubernetes/bundled.yaml"

	// Install manager
	cmd := fmt.Sprintf("kubectl apply -f %s --wait=true --timeout=300s", managerURL)
	log.Println(cmd)
	if out, err := ssh.Run(cmd, host, key, user, "", true, 60); err != nil {
		return fmt.Errorf("failed to install KubeVirt Manager: %w", err)
	} else {
		log.Println(out)
	}

	// Wait for manager to be ready
	waitCmd := "kubectl wait --for=condition=ready --timeout=300s pod -l app=kubevirt-manager -n kubevirt-manager"
	log.Println(waitCmd)
	if _, err := ssh.Run(waitCmd, host, key, user, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt Manager: %w", err)
	}

	return nil
}
