package k3s

import (
	"fmt"
	"os"
	"os/exec"
)

// InstallHelm installs Helm to k3s cluster.
func InstallHelm() error {
	command := exec.Command("curl", "-sfL", "https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("error installing Helm: %w", err)
	}
	return nil
}
