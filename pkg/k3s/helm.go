package k3s

import (
	"fmt"
	"log"

	"github.com/rusik69/govnocloud2/pkg/ssh"
)

// HelmConfig holds Helm installation configuration
type HelmConfig struct {
	ScriptURL string
	Shell     string
	Host      string
	User      string
	Key       string
}

// NewHelmConfig creates a default Helm configuration
func NewHelmConfig(host, user, key string) *HelmConfig {
	return &HelmConfig{
		ScriptURL: "https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3",
		Shell:     "bash",
		Host:      host,
		User:      user,
		Key:       key,
	}
}

// InstallHelm installs Helm to k3s cluster.
func InstallHelm(host, user, key string) error {
	cfg := NewHelmConfig(host, user, key)
	return installHelmWithConfig(cfg)
}

// installHelmWithConfig handles the actual Helm installation
func installHelmWithConfig(cfg *HelmConfig) error {
	cmd := fmt.Sprintf("curl --no-progress-meter -sfL %s | %s", cfg.ScriptURL, cfg.Shell)
	out, err := ssh.Run(cmd, cfg.Host, cfg.Key, cfg.User, "", true, 60)
	if err != nil {
		return fmt.Errorf("error installing Helm: %w", err)
	}
	log.Println(out)
	return nil
}
