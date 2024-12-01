package k3s

import (
	"fmt"
	"os"
	"os/exec"
)

// HelmConfig holds Helm installation configuration
type HelmConfig struct {
	ScriptURL string
	Shell     string
}

// NewHelmConfig creates a default Helm configuration
func NewHelmConfig() *HelmConfig {
	return &HelmConfig{
		ScriptURL: "https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3",
		Shell:     "bash",
	}
}

// InstallHelm installs Helm to k3s cluster.
func InstallHelm() error {
	cfg := NewHelmConfig()
	return installHelmWithConfig(cfg)
}

// installHelmWithConfig handles the actual Helm installation
func installHelmWithConfig(cfg *HelmConfig) error {
	cmd := fmt.Sprintf("curl -sfL %s | %s", cfg.ScriptURL, cfg.Shell)
	command := exec.Command(cfg.Shell, "-c", cmd)
	
	// Set up command output
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing Helm: %w", err)
	}
	
	return nil
}
