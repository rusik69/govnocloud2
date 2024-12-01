package k3s

import (
	"fmt"
	"os"
	"os/exec"
)

// KubeVirtConfig holds KubeVirt installation configuration
type KubeVirtConfig struct {
	Version     string
	BaseURL     string
	BinaryPath  string
	Permissions string
}

// NewKubeVirtConfig creates a default KubeVirt configuration
func NewKubeVirtConfig() *KubeVirtConfig {
	const version = "v1.3.1"
	return &KubeVirtConfig{
		Version:     version,
		BaseURL:     fmt.Sprintf("https://github.com/kubevirt/kubevirt/releases/download/%s", version),
		BinaryPath:  "/usr/local/bin/virtctl",
		Permissions: "+x",
	}
}

// InstallKubeVirt installs KubeVirt to k3s cluster.
func InstallKubeVirt() error {
	cfg := NewKubeVirtConfig()
	
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

	return nil
}

// applyKubeVirtManifest applies a KubeVirt manifest using kubectl
func applyKubeVirtManifest(cfg *KubeVirtConfig, manifest string) error {
	url := fmt.Sprintf("%s/%s", cfg.BaseURL, manifest)
	
	cmd := exec.Command("kubectl", "apply", "-f", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error applying manifest %s: %w", manifest, err)
	}
	
	return nil
}

// installVirtctl downloads and installs the virtctl binary
func installVirtctl(cfg *KubeVirtConfig) error {
	// Download virtctl
	virtctlURL := fmt.Sprintf("%s/virtctl-linux-amd64", cfg.BaseURL)
	downloadCmd := exec.Command("sudo", "curl", "-L", "-o", cfg.BinaryPath, virtctlURL)
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr
	
	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("error downloading virtctl: %w", err)
	}

	// Make virtctl executable
	chmodCmd := exec.Command("sudo", "chmod", cfg.Permissions, cfg.BinaryPath)
	chmodCmd.Stdout = os.Stdout
	chmodCmd.Stderr = os.Stderr
	
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("error setting virtctl permissions: %w", err)
	}

	return nil
}
