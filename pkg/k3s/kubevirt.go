package k3s

import (
	"fmt"
	"log"
	"time"

	"github.com/rusik69/govnocloud2/pkg/ssh"
	"github.com/rusik69/govnocloud2/pkg/types"
)

// KubeVirtConfig holds KubeVirt installation configuration
type KubeVirtConfig struct {
	Version     string
	BaseURL     string
	BinaryPath  string
	Permissions string
	Host        string
	User        string
	Key         string
}

// NewKubeVirtConfig creates a default KubeVirt configuration
func NewKubeVirtConfig(host, user, key string) *KubeVirtConfig {
	const version = "v1.4.0"
	return &KubeVirtConfig{
		Version:     version,
		BaseURL:     fmt.Sprintf("https://github.com/kubevirt/kubevirt/releases/download/%s", version),
		BinaryPath:  "/usr/local/bin/virtctl",
		Permissions: "+x",
		Host:        host,
		User:        user,
		Key:         key,
	}
}

// InstallKubeVirt installs KubeVirt to k3s cluster.
func InstallKubeVirt(host, user, key string) error {
	cfg := NewKubeVirtConfig(host, user, key)

	// Install operator
	if err := applyKubeVirtManifest(cfg, "kubevirt-operator.yaml"); err != nil {
		return fmt.Errorf("failed to install KubeVirt operator: %w", err)
	}

	// Install CR
	if err := applyKubeVirtManifest(cfg, "kubevirt-cr.yaml"); err != nil {
		return fmt.Errorf("failed to install KubeVirt CR: %w", err)
	}

	// Install virtctl
	if err := installVirtctl(cfg); err != nil {
		return fmt.Errorf("failed to install virtctl: %w", err)
	}

	log.Println("sleeping 5 seconds")
	time.Sleep(5 * time.Second)
	log.Println("Waiting for KubeVirt to be ready")
	// Wait for KubeVirt to be ready
	cmd := "kubectl wait --for=condition=ready --timeout=300s pod -l kubevirt.io=virt-operator -n kubevirt"
	log.Println(cmd)
	if _, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 300); err != nil {
		return fmt.Errorf("failed to wait for KubeVirt to be ready: %w", err)
	}
	// FIXME: wait for crds
	time.Sleep(5 * time.Second)
	// Create kubevirt instance types
	for _, size := range types.VMSizes {
		cmd := fmt.Sprintf("virtctl create instancetype --name %s --cpu %d --memory %dMi | kubectl create -f -", size.Name, size.CPU, size.RAM)
		log.Println(cmd)
		if _, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60); err != nil {
			return fmt.Errorf("failed to create kubevirt instance type: %w", err)
		}
	}
	return nil
}

// applyKubeVirtManifest applies a KubeVirt manifest using kubectl
func applyKubeVirtManifest(cfg *KubeVirtConfig, manifest string) error {
	url := fmt.Sprintf("%s/%s", cfg.BaseURL, manifest)
	cmd := fmt.Sprintf("kubectl apply -f %s", url)
	log.Println(cmd)
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to apply KubeVirt manifest: %w", err)
	}
	log.Println(out)

	return nil
}

// installVirtctl downloads and installs the virtctl binary
func installVirtctl(cfg *KubeVirtConfig) error {
	// Download virtctl
	virtctlURL := fmt.Sprintf("%s/virtctl-v1.4.0-linux-amd64", cfg.BaseURL)
	cmd := fmt.Sprintf("sudo curl -L -o %s %s; sudo chmod +x %s", cfg.BinaryPath, virtctlURL, cfg.BinaryPath)
	log.Println(cmd)
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("failed to download virtctl: %w", err)
	}
	log.Println(out)
	return nil
}
