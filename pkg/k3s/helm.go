package k3s

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallHelm installs Helm to k3s cluster.
func InstallHelm() error {
	cmd := "curl -sfL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
	command := exec.Command("bash", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing Helm: %w", err)
	}
	return nil
}
